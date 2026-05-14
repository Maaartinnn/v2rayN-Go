package subscription

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"v2rayn-go/database"
	"v2rayn-go/parser"
)

// Service 订阅管理服务
type Service struct {
	client *http.Client
}

// NewService 创建订阅服务
func NewService() *Service {
	return &Service{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// UpdateGroupSubscription 更新单个订阅分组
func (s *Service) UpdateGroupSubscription(group *database.NodeGroup, useProxy bool) error {
	if !group.IsSubscription || group.URL == "" {
		return fmt.Errorf("group %s is not a subscription group", group.Alias)
	}

	log.Printf("Updating group subscription: %s (%s)", group.Alias, group.URL)

	// 创建 HTTP 客户端
	client := s.client
	if useProxy {
		client = &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		}
	}

	// 拉取订阅内容
	content, err := s.fetchContentWithClient(client, group.URL, group.UserAgent)
	if err != nil {
		return fmt.Errorf("failed to fetch subscription: %w", err)
	}

	// 解析订阅内容
	profiles, err := parser.ParseSubscriptionContent(content)
	if err != nil {
		return fmt.Errorf("failed to parse subscription content: %w", err)
	}

	// 应用别名正则过滤
	if group.AliasRegex != "" {
		profiles = filterProfilesByAlias(profiles, group.AliasRegex)
	}

	// 开始数据库事务
	tx := database.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// 删除该分组下的旧节点
	if err := tx.Where("group_uuid = ?", group.UUID).Delete(&database.Profile{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete old profiles: %w", err)
	}

	// 插入新节点（步长排序）
	seq := database.SortSequence(len(profiles))
	for i, profile := range profiles {
		profile.GroupUUID = group.UUID
		profile.SortOrder = seq[i]
		if err := tx.Create(profile).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create profile: %w", err)
		}
	}

	// 更新分组的最后更新时间和节点数
	now := time.Now()
	group.LastUpdateTime = now
	group.NodeCount = len(profiles)
	if err := tx.Save(group).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update group: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Group %s updated: %d profiles", group.Alias, len(profiles))
	return nil
}

// UpdateGroupByID 根据 ID 更新单个订阅分组
func (s *Service) UpdateGroupByID(groupID uint, useProxy bool) error {
	var group database.NodeGroup
	if err := database.DB.First(&group, groupID).Error; err != nil {
		return fmt.Errorf("group not found: %w", err)
	}
	return s.UpdateGroupSubscription(&group, useProxy)
}

// UpdateAllSubscriptions 更新所有启用的订阅分组
func (s *Service) UpdateAllSubscriptions() {
	var groups []database.NodeGroup
	if err := database.DB.Where("is_subscription = ? AND enable_update = ? AND enabled = ?", true, true, true).Find(&groups).Error; err != nil {
		log.Printf("Failed to query subscription groups: %v", err)
		return
	}

	if len(groups) == 0 {
		log.Println("No enabled subscription groups found")
		return
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(groups))

	for i := range groups {
		wg.Add(1)
		go func(g *database.NodeGroup) {
			defer wg.Done()
			if err := s.UpdateGroupSubscription(g, false); err != nil {
				errChan <- fmt.Errorf("group %s: %w", g.Alias, err)
			}
		}(&groups[i])
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		log.Printf("Subscription update error: %v", err)
	}
}

// StartAutoUpdate 启动自动更新定时器
func (s *Service) StartAutoUpdate(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// 启动时立即执行一次
	s.UpdateAllSubscriptions()

	for {
		select {
		case <-ctx.Done():
			log.Println("Auto update stopped")
			return
		case <-ticker.C:
			s.checkAndUpdateSubscriptions()
		}
	}
}

// checkAndUpdateSubscriptions 检查并更新到期的订阅
func (s *Service) checkAndUpdateSubscriptions() {
	var groups []database.NodeGroup
	if err := database.DB.Where("is_subscription = ? AND enable_update = ? AND enabled = ?", true, true, true).Find(&groups).Error; err != nil {
		log.Printf("Failed to query subscription groups: %v", err)
		return
	}

	now := time.Now()
	for i := range groups {
		g := &groups[i]
		if g.UpdateInterval <= 0 {
			continue
		}
		nextUpdate := g.LastUpdateTime.Add(time.Duration(g.UpdateInterval) * time.Minute)
		if now.After(nextUpdate) {
			if err := s.UpdateGroupSubscription(g, false); err != nil {
				log.Printf("Failed to update group %s: %v", g.Alias, err)
			}
		}
	}
}

// fetchContentWithClient 使用指定客户端拉取订阅内容
func (s *Service) fetchContentWithClient(client *http.Client, rawURL string, userAgent string) (string, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return "", err
	}

	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	} else {
		req.Header.Set("User-Agent", "v2rayN-Go/1.0")
	}
	req.Header.Set("Accept", "*/*")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(body)), nil
}

// fetchContent 拉取订阅内容（使用默认客户端）
func (s *Service) fetchContent(rawURL string) (string, error) {
	return s.fetchContentWithClient(s.client, rawURL, "")
}

// filterProfilesByAlias 根据正则表达式过滤节点别名
func filterProfilesByAlias(profiles []*database.Profile, pattern string) []*database.Profile {
	re, err := regexp.Compile(pattern)
	if err != nil {
		log.Printf("Invalid alias regex: %s, error: %v", pattern, err)
		return profiles
	}

	var filtered []*database.Profile
	for _, p := range profiles {
		if re.MatchString(p.Name) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// ParseURLForProxy 解析 URL 并应用代理设置
func ParseURLForProxy(rawURL string) (*url.URL, error) {
	return url.Parse(rawURL)
}
