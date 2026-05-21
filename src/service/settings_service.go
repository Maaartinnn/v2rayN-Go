package service

import (
	"fmt"

	"v2rayn-go/config"
)

// SettingsService 配置管理业务逻辑层
type SettingsService struct {
	cfg *config.AppConfig
}

// NewSettingsService 创建配置服务
func NewSettingsService(cfg *config.AppConfig) *SettingsService {
	return &SettingsService{cfg: cfg}
}

// GetSettings 获取当前配置
func (s *SettingsService) GetSettings() map[string]any {
	return map[string]any{
		"listen_ip":     s.cfg.ListenIP,
		"web_port":      s.cfg.WebPort,
		"socks_port":    s.cfg.SocksPort,
		"http_port":     s.cfg.HTTPPort,
		"outbound_ip":   s.cfg.OutboundIP,
		"github_mirror": s.cfg.GitHubMirror,
	}
}

// UpdateSettingsRequest 配置更新请求
type UpdateSettingsRequest struct {
	ListenIP     *string `json:"listen_ip"`
	SocksPort    *int    `json:"socks_port"`
	HTTPPort     *int    `json:"http_port"`
	OutboundIP   *string `json:"outbound_ip"`
	GitHubMirror *string `json:"github_mirror"`
}

// UpdateSettings 更新配置
func (s *SettingsService) UpdateSettings(req *UpdateSettingsRequest) error {
	if req.ListenIP != nil {
		s.cfg.ListenIP = *req.ListenIP
	}
	if req.SocksPort != nil && *req.SocksPort > 0 {
		s.cfg.SocksPort = *req.SocksPort
	}
	if req.HTTPPort != nil && *req.HTTPPort > 0 {
		s.cfg.HTTPPort = *req.HTTPPort
	}
	if req.OutboundIP != nil {
		s.cfg.OutboundIP = *req.OutboundIP
	}
	if req.GitHubMirror != nil {
		s.cfg.GitHubMirror = *req.GitHubMirror
	}
	if err := s.cfg.SaveJSONConfig(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}
