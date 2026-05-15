package service

import (
	"context"

	"v2rayn-go/database"
	"v2rayn-go/subscription"
)

// PingServiceInterface 是 ping 服务的接口
type PingServiceInterface interface {
	PingAllProfiles(ctx context.Context, concurrency int) []subscription.PingResult
	PingSingleProfile(profile *database.Profile) subscription.PingResult
}

// NewPingService 创建 ping 服务
func NewPingService() PingServiceInterface {
	return subscription.NewPingService()
}
