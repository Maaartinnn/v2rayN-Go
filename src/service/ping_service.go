package service

import (
	"context"

	"v2rayn-go/database"
	"v2rayn-go/ping"
)

// PingServiceInterface 是 ping 服务的接口
type PingServiceInterface interface {
	PingAllProfiles(ctx context.Context, concurrency int) []ping.PingResult
	PingSingleProfile(profile *database.Profile) ping.PingResult
}

// NewPingService 创建 ping 服务
func NewPingService() PingServiceInterface {
	return ping.NewPingService()
}
