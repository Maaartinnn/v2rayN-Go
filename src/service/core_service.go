package service

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"v2rayn-go/config"
	"v2rayn-go/configbuilder"
	"v2rayn-go/core"
	"v2rayn-go/coredef"
	"v2rayn-go/database"
	"v2rayn-go/updater"
)

// ProgressCallback 下载进度回调
type ProgressCallback func(downloaded, total int64)

// CoreService 核心管理业务逻辑层
type CoreService struct {
	cfg     *config.AppConfig
	coreMgr *core.CoreAdminManager
	updater *updater.Updater
}

// NewCoreService 创建核心服务
func NewCoreService(cfg *config.AppConfig, coreMgr *core.CoreAdminManager) *CoreService {
	return &CoreService{
		cfg:     cfg,
		coreMgr: coreMgr,
		updater: updater.NewUpdater(cfg),
	}
}

// Start 启动核心。如果 configPath 为空，自动查询活跃节点和路由规则来生成配置。
func (s *CoreService) Start(coreType string, configPath string) error {
	if configPath == "" {
		var err error
		configPath, coreType, err = s.buildConfig(coreType)
		if err != nil {
			return err
		}
	}

	if err := s.coreMgr.StartCore(core.CoreType(coreType), configPath); err != nil {
		return fmt.Errorf("failed to start core: %w", err)
	}
	return nil
}

// buildConfig 根据活跃节点和路由规则生成配置文件
func (s *CoreService) buildConfig(coreType string) (configPath string, resolvedCoreType string, err error) {
	var profile database.Profile
	if err := database.DB.Where("is_active = ?", true).First(&profile).Error; err != nil {
		return "", "", fmt.Errorf("no active profile selected")
	}

	if coreType == "" {
		coreType = profile.CoreType
	}

	var rules []database.RoutingRule
	if err := database.DB.Order("sort_order ASC").Find(&rules).Error; err != nil {
		return "", "", fmt.Errorf("failed to load routing rules: %w", err)
	}

	switch coreType {
	case "xray":
		configPath, err = configbuilder.SaveXrayConfig(&profile, rules, s.cfg.AppDir, s.cfg.SocksPort, s.cfg.HTTPPort)
	case "sing-box":
		configPath, err = configbuilder.SaveSingboxConfig(&profile, rules, s.cfg.AppDir, s.cfg.SocksPort)
	default:
		return "", "", fmt.Errorf("unsupported core type: %s", coreType)
	}

	if err != nil {
		return "", "", fmt.Errorf("failed to build config: %w", err)
	}
	return configPath, coreType, nil
}

// Stop 停止核心
func (s *CoreService) Stop(coreType string) error {
	if coreType == "" {
		coreType = "xray"
	}
	if err := s.coreMgr.StopCore(core.CoreType(coreType)); err != nil {
		return fmt.Errorf("failed to stop core: %w", err)
	}
	return nil
}

// GetAllStatus 获取所有核心状态
func (s *CoreService) GetAllStatus() []core.CoreInfo {
	return s.coreMgr.GetAllStatus()
}

// GetLocalCores 获取本地核心列表
func (s *CoreService) GetLocalCores() []updater.CoreInfo {
	return s.updater.GetLocalCores()
}

// CheckUpdates 检查所有核心更新，返回 latestVersions map
func (s *CoreService) CheckUpdates() map[string]string {
	cores := s.updater.CheckAllUpdates()
	latestVersions := make(map[string]string)
	for _, c := range cores {
		if c.LatestVer != "" {
			ver := strings.TrimPrefix(c.LatestVer, "v")
			latestVersions[c.Name] = ver
		}
	}
	return latestVersions
}

// DetectVersions 异步检测本地核心版本，完成后调用 onComplete 回调
func (s *CoreService) DetectVersions(onComplete func(map[string]string)) {
	go func() {
		cores := s.updater.GetLocalCoresWithVersions()
		versions := make(map[string]string)
		for _, c := range cores {
			if c.Version != "" {
				versions[c.Name] = c.Version
			}
		}
		onComplete(versions)
	}()
}

// Download 下载核心（使用默认源）
func (s *CoreService) Download(name string, onProgress ProgressCallback) error {
	return s.updater.DownloadCore(name, onProgress)
}

// DownloadFromURL 从指定 URL 下载核心
func (s *CoreService) DownloadFromURL(name, downloadURL string, onProgress ProgressCallback) error {
	return s.updater.DownloadCoreFromURL(name, downloadURL, onProgress)
}

// Upload 上传核心二进制文件（支持普通文件和压缩包）
func (s *CoreService) Upload(coreName string, filename string, data io.Reader) (string, error) {
	coreType := coredef.CoreType(coreName)
	meta, exists := coredef.Registry[coreType]
	if !exists {
		return "", fmt.Errorf("unsupported core: %s", coreName)
	}

	coreDir := filepath.Join(s.cfg.BinDir, meta.SubDir)
	if err := os.MkdirAll(coreDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create core directory: %w", err)
	}

	destPath := filepath.Join(coreDir, meta.BinaryName())

	lowerName := strings.ToLower(filename)
	isArchive := strings.HasSuffix(lowerName, ".zip") || strings.HasSuffix(lowerName, ".tar.gz") || strings.HasSuffix(lowerName, ".tgz")

	if isArchive {
		tmpFile, err := os.CreateTemp("", "v2rayn-upload-*.tmp")
		if err != nil {
			return "", fmt.Errorf("failed to create temp file: %w", err)
		}
		tmpPath := tmpFile.Name()
		defer os.Remove(tmpPath)
		if _, err := io.Copy(tmpFile, data); err != nil {
			tmpFile.Close()
			return "", fmt.Errorf("failed to save temp file: %w", err)
		}
		tmpFile.Close()
		if err := s.updater.ExtractBinary(tmpPath, filename, destPath, meta.BinaryName()); err != nil {
			return "", fmt.Errorf("failed to extract binary from archive: %w", err)
		}
	} else {
		dst, err := os.Create(destPath)
		if err != nil {
			return "", fmt.Errorf("failed to create file: %w", err)
		}
		defer dst.Close()
		if _, err := io.Copy(dst, data); err != nil {
			return "", fmt.Errorf("failed to save file: %w", err)
		}
	}

	if runtime.GOOS != "windows" {
		os.Chmod(destPath, 0755)
	}

	log.Printf("Uploaded core: %s (%s)", coreName, filename)
	return destPath, nil
}
