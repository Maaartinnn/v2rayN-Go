package subscription

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
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

// UpdateSubscription 更新单个订阅
func (s *Service) UpdateSubscription(sub *database.Subscription) error {
	log.Printf("Updating subscription: %s (%s)", sub.Name, sub.URL)

	// 拉取订阅内容
	content, err := s.fetchContent(sub.URL)
	if err != nil {
		return fmt.Errorf("failed to fetch subscription: %w", err)
	}

	// 解析订阅内容
	profiles, err := parser.ParseSubscriptionContent(content)
	if err != nil {
		return fmt.Errorf("failed to parse subscription content: %w", err)
	}

	// 开始数据库事务
	tx := database.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// 删除该订阅下的旧节点
	if err := tx.Where("subscription_id = ?", sub.ID).Delete(&database.Profile{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete old profiles: %w", err)
	}

	// 插入新节点
	for i, profile := range profiles {
		profile.SubscriptionID = sub.ID
		profile.GroupName = sub.Name
		profile.SortOrder = i
		if err := tx.Create(profile).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create profile: %w", err)
		}
	}

	// 更新订阅的最后更新时间
	now := time.Now()
	sub.LastUpdateTime = now
	if err := tx.Save(sub).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Subscription %s updated: %d profiles", sub.Name, len(profiles))
	return nil
}

// UpdateAllSubscriptions 更新所有启用的订阅
func (s *Service) UpdateAllSubscriptions() {
	var subs []database.Subscription
	if err := database.DB.Where("enabled = ?", true).Find(&subs).Error; err != nil {
		log.Printf("Failed to query subscriptions: %v", err)
		return
	}

	if len(subs) == 0 {
		log.Println("No enabled subscriptions found")
		return
	}

	// 使用 errgroup 并发更新
	var wg sync.WaitGroup
	errChan := make(chan error, len(subs))

	for i := range subs {
		wg.Add(1)
		go func(sub *database.Subscription) {
			defer wg.Done()
			if err := s.UpdateSubscription(sub); err != nil {
				errChan <- fmt.Errorf("subscription %s: %w", sub.Name, err)
			}
		}(&subs[i])
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		log.Printf("Subscription update error: %v", err)
	}
}

// StartAutoUpdate 启动自动更新定时器
func (s *Service) StartAutoUpdate(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
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
	var subs []database.Subscription
	if err := database.DB.Where("enabled = ? AND auto_update = ?", true, true).Find(&subs).Error; err != nil {
		log.Printf("Failed to query subscriptions: %v", err)
		return
	}

	now := time.Now()
	for i := range subs {
		sub := &subs[i]
		// 检查是否到达更新间隔
		if sub.UpdateInterval <= 0 {
			sub.UpdateInterval = 86400 // 默认 24 小时
		}
		nextUpdate := sub.LastUpdateTime.Add(time.Duration(sub.UpdateInterval) * time.Second)
		if now.After(nextUpdate) {
			if err := s.UpdateSubscription(sub); err != nil {
				log.Printf("Failed to update subscription %s: %v", sub.Name, err)
			}
		}
	}
}

// fetchContent 拉取订阅内容
func (s *Service) fetchContent(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	// 设置常见请求头
	req.Header.Set("User-Agent", "v2rayN-Go/1.0")
	req.Header.Set("Accept", "*/*")

	resp, err := s.client.Do(req)
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
