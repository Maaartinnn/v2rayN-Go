package ping

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"v2rayn-go/database"
	"v2rayn-go/httpclient"

	"gorm.io/gorm"
)

// PingResult 测速结果
type PingResult struct {
	ProfileUUID string `json:"profile_uuid"`
	Latency     int    `json:"latency"` // 毫秒
	Error       string `json:"error,omitempty"`
}

// PingService 延迟测速服务
type PingService struct {
	client *http.Client
}

// NewPingService 创建测速服务
func NewPingService() *PingService {
	return &PingService{
		client: httpclient.NewClient(10 * time.Second),
	}
}

// TCPPing TCP 连接延迟测试
func (ps *PingService) TCPPing(host string, port int) (int, error) {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	start := time.Now()

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return 0, fmt.Errorf("TCP connect failed: %w", err)
	}
	conn.Close()

	latency := time.Since(start).Milliseconds()
	return int(latency), nil
}

// HTTPPing HTTP 真连通性测试
func (ps *PingService) HTTPPing(url string) (int, error) {
	start := time.Now()

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, err
	}
	// User-Agent 由 httpclient.Transport 自动注入

	resp, err := ps.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	latency := time.Since(start).Milliseconds()
	return int(latency), nil
}

// PingProfile 测试单个节点延迟（TCP Ping）
func (ps *PingService) PingProfile(profile *database.Profile) PingResult {
	latency, err := ps.TCPPing(profile.ProxyAddress, profile.ProxyPort)
	if err != nil {
		return PingResult{
			ProfileUUID: profile.UUID,
			Latency:     0,
			Error:       err.Error(),
		}
	}

	return PingResult{
		ProfileUUID: profile.UUID,
		Latency:     latency,
	}
}

// PingProfiles 批量测速节点（并发）
func (ps *PingService) PingProfiles(profiles []database.Profile, concurrency int) []PingResult {
	if concurrency <= 0 {
		concurrency = 10
	}

	results := make([]PingResult, len(profiles))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i := range profiles {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int) {
			defer wg.Done()
			defer func() { <-sem }()
			results[idx] = ps.PingProfile(&profiles[idx])
		}(i)
	}

	wg.Wait()
	return results
}

// PingSingleProfile 测试单个节点延迟并更新数据库
func (ps *PingService) PingSingleProfile(profile *database.Profile) PingResult {
	result := ps.PingProfile(profile)

	// 提取公共变量，消除 if/else 冗余
	testResult := fmt.Sprintf("%dms", result.Latency)
	if result.Error != "" {
		testResult = "timeout"
	}
	now := time.Now()

	database.DB.Model(&database.Profile{}).
		Where("uuid = ?", result.ProfileUUID).
		Updates(map[string]interface{}{
			"test_result":    testResult,
			"last_test_time": now,
		})

	return result
}

// PingAllProfiles 测试数据库中所有节点的延迟
func (ps *PingService) PingAllProfiles(ctx context.Context, concurrency int) []PingResult {
	var profiles []database.Profile
	if err := database.DB.Find(&profiles).Error; err != nil {
		log.Printf("Failed to query profiles: %v", err)
		return nil
	}

	if len(profiles) == 0 {
		log.Println("No profiles to ping")
		return nil
	}

	log.Printf("Starting latency test for %d profiles (concurrency: %d)", len(profiles), concurrency)
	results := ps.PingProfiles(profiles, concurrency)

	// 统一获取当前时间，保证同一批次更新的时间戳完全一致
	now := time.Now()

	// 使用显式事务包裹整个批量更新，将 N 次隐式事务合并为 1 次，
	// 对 SQLite 而言将 N 次 fsync 降为 1 次，性能提升数量级。
	txErr := database.DB.Transaction(func(tx *gorm.DB) error {
		for _, result := range results {
			testResult := fmt.Sprintf("%dms", result.Latency)
			if result.Error != "" {
				testResult = "timeout"
			}

			if err := tx.Model(&database.Profile{}).
				Where("uuid = ?", result.ProfileUUID).
				Updates(map[string]interface{}{
					"test_result":    testResult,
					"last_test_time": now,
				}).Error; err != nil {
				return fmt.Errorf("update profile %s failed: %w", result.ProfileUUID, err)
			}
		}
		return nil
	})

	if txErr != nil {
		log.Printf("[ERROR] Batch update ping results failed: %v", txErr)
	}

	// 统计结果
	success := 0
	for _, r := range results {
		if r.Error == "" {
			success++
		}
	}
	log.Printf("Latency test completed: %d/%d successful", success, len(results))

	return results
}
