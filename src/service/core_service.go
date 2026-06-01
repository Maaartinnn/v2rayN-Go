package service

import (
	"fmt"
	"io"
	"log"
	"log/slog"
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
//
// 默认使用 stdin 无文件落地模式：配置数据直接通过 cmd.Stdin 管道注入内核进程，
// 全程不触碰物理磁盘。如果 CoreConfigDebug 为 true，则写入配置文件并使用文件模式启动。
func (s *CoreService) Start(coreType string, configPath string) error {
	// 用户手动指定了配置文件路径 → 传统文件启动
	if configPath != "" {
		if err := s.coreMgr.StartCore(core.CoreType(coreType), configPath); err != nil {
			return fmt.Errorf("failed to start core: %w", err)
		}
		return nil
	}

	// 查询活跃节点和路由规则
	var profile database.Profile
	if err := database.DB.Where("is_active = ?", true).First(&profile).Error; err != nil {
		return NewValidation("no active profile selected", nil)
	}

	if coreType == "" {
		coreType = profile.CoreType
	}

	// 节点未显式指定内核类型时，自动选择协议兼容且已安装的最佳内核
	if coreType == "" {
		installedCores := s.GetCompatibleInstalledCores(profile.ProxyProtocol)
		if len(installedCores) > 0 {
			coreType = installedCores[0] // 第一个是推荐优先级最高的
		} else {
			coreType = "xray" // 极端兜底：没有任何兼容内核安装
		}
	}

	var rules []database.RoutingRule
	if err := database.DB.Order("sort_order ASC").Find(&rules).Error; err != nil {
		return fmt.Errorf("failed to load routing rules: %w", err)
	}

	builder, ok := configbuilder.GetBuilder(coreType)
	if !ok {
		return NewValidation("unsupported core type: "+coreType, nil)
	}

	params := &configbuilder.BuildConfigParams{
		Profile:   &profile,
		Rules:     rules,
		ConfigDir: s.cfg.AppDir,
		SocksPort: s.cfg.SocksPort,
		HTTPPort:  s.cfg.HTTPPort,
	}

	if s.cfg.CoreConfigDebug {
		// 调试模式：写入文件，传统启动
		configPath, err := builder.Build(params)
		if err != nil {
			return fmt.Errorf("failed to build config: %w", err)
		}
		slog.Info("debug mode: kernel config saved to disk", "path", configPath)
		if err := s.coreMgr.StartCore(core.CoreType(coreType), configPath); err != nil {
			return fmt.Errorf("failed to start core: %w", err)
		}
		return nil
	}

	// 生产模式：stdin 无文件落地
	configData, err := builder.BuildBytes(params)
	if err != nil {
		return fmt.Errorf("failed to build config: %w", err)
	}
	slog.Info("stdin mode: config injected via pipe", "core", coreType, "size", len(configData))
	if err := s.coreMgr.StartCore(core.CoreType(coreType), "", core.WithStdin(configData)); err != nil {
		return fmt.Errorf("failed to start core: %w", err)
	}
	return nil
}

// GetCompatibleInstalledCores 获取协议兼容且本地已安装的内核列表（按推荐优先级排序）。
//
// 此方法是"协议→内核"选择的唯一后端逻辑，被以下两处复用：
//   - NodeEditForm 编辑页面：通过 GET /api/profiles/{uuid} 返回 core_list 供前端展示
//   - CoreService.Start()：当用户未指定 coreType 时，选择最佳默认内核
//
// 算法：
//  1. 根据 ProtocolCoreMap 获取协议支持的内核列表（已按推荐优先级排序）
//  2. 检查 bin/ 目录下每个内核的二进制文件是否已安装
//  3. 只返回既兼容又已安装的内核（保持推荐优先级顺序）
//
// 注意：返回值确保是 []string (空数组) 而不是 nil，避免 JSON 序列化为 null
func (s *CoreService) GetCompatibleInstalledCores(protocol string) []string {
	supported := coredef.GetSupportedCoresForProtocol(protocol)
	if len(supported) == 0 {
		return make([]string, 0) // 返回空数组而非 nil
	}

	// 获取本地已安装的内核集合
	localCores := s.updater.GetLocalCores()
	installedSet := make(map[string]bool)
	for _, c := range localCores {
		if c.Version == "installed" {
			installedSet[c.Name] = true
		}
	}

	// 按推荐优先级过滤，只保留已安装的
	// 初始化为空数组，确保 JSON 序列化为 []
	result := make([]string, 0)
	for _, ct := range supported {
		if installedSet[string(ct)] {
			result = append(result, string(ct))
		}
	}
	return result
}

// GetInstalledCoreMatrix 返回当前环境所有协议对应的可用内核矩阵。
//
// 用于前端一次性获取所有协议的兼容性数据，避免前端在切换协议时反复请求后端。
// 返回格式：{"vmess": ["xray", "sing-box"], "anytls": ["sing-box"]}
func (s *CoreService) GetInstalledCoreMatrix() map[string][]string {
	matrix := make(map[string][]string)
	for protocol := range coredef.ProtocolCoreMap {
		// 复用 GetCompatibleInstalledCores 保证一致性
		cores := s.GetCompatibleInstalledCores(protocol)
		// 仅当该协议有可用内核时才加入矩阵
		if len(cores) > 0 {
			matrix[protocol] = cores
		}
	}
	return matrix
}

// Stop 停止指定类型的内核。
//
// coreType 为空时，自动停止所有运行中的内核（用户点击 HomeView 开关按钮时
// 不需要知道当前运行的是哪个内核，直接全部停止即可）。
func (s *CoreService) Stop(coreType string) error {
	if coreType == "" {
		s.coreMgr.StopAll()
		return nil
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
		return "", NewValidation("unsupported core: "+coreName, nil)
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
