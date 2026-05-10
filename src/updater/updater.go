package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"v2rayn-go/config"

	"golang.org/x/sys/cpu"
)

// CoreInfo 内核信息
type CoreInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Repo        string `json:"repo"`           // GitHub 仓库，如 "XTLS/Xray-core"
	Version     string `json:"version"`        // 当前安装版本
	LatestVer   string `json:"latest_version"` // 最新可用版本
	BinaryName  string `json:"binary_name"`    // 可执行文件名
	SubDir      string `json:"sub_dir"`        // 嵌套子目录名
}

// GitHubRelease GitHub Release API 响应
type GitHubRelease struct {
	TagName string        `json:"tag_name"`
	Name    string        `json:"name"`
	Assets  []GitHubAsset `json:"assets"`
}

// GitHubAsset GitHub Release Asset
type GitHubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// releaseCacheEntry release 缓存条目
type releaseCacheEntry struct {
	release   *GitHubRelease
	fetchedAt time.Time
}

// releaseCacheTTL release 缓存有效期
const releaseCacheTTL = 300 * time.Second

// Updater 内核更新管理器
type Updater struct {
	cfg          *config.AppConfig
	client       *http.Client
	releaseCache map[string]*releaseCacheEntry // repo -> 缓存条目
	cacheMu      sync.RWMutex                  // 保护 releaseCache 的读写锁
}

// NewUpdater 创建更新管理器
func NewUpdater(cfg *config.AppConfig) *Updater {
	return &Updater{
		cfg: cfg,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
		releaseCache: make(map[string]*releaseCacheEntry),
	}
}

// GetSupportedCores 获取支持的内核列表
func (u *Updater) GetSupportedCores() []CoreInfo {
	return []CoreInfo{
		{
			Name:        "xray",
			DisplayName: "Xray-core",
			Repo:        "XTLS/Xray-core",
			BinaryName:  getBinaryName("xray"),
			SubDir:      "xray",
		},
		{
			Name:        "sing-box",
			DisplayName: "Sing-box",
			Repo:        "SagerNet/sing-box",
			BinaryName:  getBinaryName("sing-box"),
			SubDir:      "sing_box",
		},
		{
			Name:        "mihomo",
			DisplayName: "Mihomo",
			Repo:        "MetaCubeX/mihomo",
			BinaryName:  getBinaryName("mihomo"),
			SubDir:      "mihomo",
		},
	}
}

// GetCoreDir 获取内核的嵌套目录路径 (bin/xray/, bin/sing_box/, bin/mihomo/)
func (u *Updater) GetCoreDir(subDir string) string {
	return filepath.Join(u.cfg.BinDir, subDir)
}

// GetCoreBinaryPath 获取内核可执行文件完整路径
func (u *Updater) GetCoreBinaryPath(coreName string) string {
	cores := u.GetSupportedCores()
	for _, c := range cores {
		if c.Name == coreName {
			return filepath.Join(u.cfg.BinDir, c.SubDir, c.BinaryName)
		}
	}
	return filepath.Join(u.cfg.BinDir, coreName)
}

// CheckUpdate 检查指定内核的最新版本
func (u *Updater) CheckUpdate(coreName string) (*CoreInfo, error) {
	cores := u.GetSupportedCores()
	var coreInfo *CoreInfo
	for i := range cores {
		if cores[i].Name == coreName {
			coreInfo = &cores[i]
			break
		}
	}
	if coreInfo == nil {
		return nil, fmt.Errorf("unsupported core: %s", coreName)
	}

	// 检查当前安装版本（嵌套目录）
	binPath := filepath.Join(u.cfg.BinDir, coreInfo.SubDir, coreInfo.BinaryName)
	if _, err := os.Stat(binPath); err == nil {
		coreInfo.Version = "installed"
	}

	// 获取最新版本
	release, err := u.getLatestRelease(coreInfo.Repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest release: %w", err)
	}
	coreInfo.LatestVer = release.TagName

	return coreInfo, nil
}

// CheckAllUpdates 检查所有内核的更新
func (u *Updater) CheckAllUpdates() []CoreInfo {
	cores := u.GetSupportedCores()
	for i := range cores {
		binPath := filepath.Join(u.cfg.BinDir, cores[i].SubDir, cores[i].BinaryName)
		if _, err := os.Stat(binPath); err == nil {
			cores[i].Version = "installed"
		}

		release, err := u.getLatestRelease(cores[i].Repo)
		if err != nil {
			log.Printf("Failed to check update for %s: %v", cores[i].Name, err)
			continue
		}
		cores[i].LatestVer = release.TagName
	}
	return cores
}

// GetLocalCores 获取所有内核的本地信息（不访问网络，不执行二进制命令，毫秒级响应）
func (u *Updater) GetLocalCores() []CoreInfo {
	cores := u.GetSupportedCores()
	for i := range cores {
		binPath := filepath.Join(u.cfg.BinDir, cores[i].SubDir, cores[i].BinaryName)
		if _, err := os.Stat(binPath); err == nil {
			cores[i].Version = "installed" // 先标记为已安装，版本号异步获取
		}
	}
	return cores
}

// GetLocalCoresWithVersions 获取所有内核的本地信息（包括版本号，需要执行二进制命令）
func (u *Updater) GetLocalCoresWithVersions() []CoreInfo {
	cores := u.GetSupportedCores()
	for i := range cores {
		binPath := filepath.Join(u.cfg.BinDir, cores[i].SubDir, cores[i].BinaryName)
		if _, err := os.Stat(binPath); err == nil {
			cores[i].Version = u.GetInstalledVersion(cores[i].Name)
		}
	}
	return cores
}

// getCoreVersionArgs 获取每个内核正确的版本查询参数
func getCoreVersionArgs(coreName string) [][]string {
	switch coreName {
	case "xray":
		return [][]string{{"version"}}
	case "sing-box":
		return [][]string{{"version"}}
	case "mihomo":
		return [][]string{{"-v"}}
	default:
		return [][]string{{"version"}, {"--version"}, {"-v"}}
	}
}

// GetInstalledVersion 获取已安装内核的版本号
func (u *Updater) GetInstalledVersion(coreName string) string {
	cores := u.GetSupportedCores()
	var coreInfo *CoreInfo
	for i := range cores {
		if cores[i].Name == coreName {
			coreInfo = &cores[i]
			break
		}
	}
	if coreInfo == nil {
		return ""
	}

	binPath := filepath.Join(u.cfg.BinDir, coreInfo.SubDir, coreInfo.BinaryName)
	if _, err := os.Stat(binPath); err != nil {
		return ""
	}

	// 使用每个内核正确的版本参数，并设置超时
	versionArgs := getCoreVersionArgs(coreName)

	for _, args := range versionArgs {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		cmd := exec.CommandContext(ctx, binPath, args...)
		output, err := cmd.CombinedOutput()
		cancel()
		if err != nil {
			continue
		}
		version := parseVersionFromOutput(string(output))
		if version != "" {
			return version
		}
	}

	// 无法获取版本号，返回 "installed"
	return "installed"
}

// versionRegex 匹配版本号：至少 X.Y 格式，可选 v 前缀
var versionRegex = regexp.MustCompile(`v?(\d+\.\d+[\.\d]*)`)

// parseVersionFromOutput 从命令输出中解析版本号（统一去掉 v 前缀）
func parseVersionFromOutput(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// 使用正则匹配版本号，支持有/无 v 前缀
		matches := versionRegex.FindStringSubmatch(line)
		if len(matches) >= 2 {
			return matches[1] // 已经不含 v 前缀
		}
	}
	return ""
}

// DownloadCore 下载指定内核的最新版本（支持镜像降级 + mihomo 版本降级）
func (u *Updater) DownloadCore(coreName string, progressFn func(downloaded, total int64)) error {
	cores := u.GetSupportedCores()
	var coreInfo *CoreInfo
	for i := range cores {
		if cores[i].Name == coreName {
			coreInfo = &cores[i]
			break
		}
	}
	if coreInfo == nil {
		return fmt.Errorf("unsupported core: %s", coreName)
	}

	// 确保内核子目录存在
	coreDir := filepath.Join(u.cfg.BinDir, coreInfo.SubDir)
	if err := os.MkdirAll(coreDir, 0755); err != nil {
		return fmt.Errorf("failed to create core directory: %w", err)
	}

	// 获取最新 release（已支持镜像降级）
	release, err := u.getLatestRelease(coreInfo.Repo)
	if err != nil {
		return fmt.Errorf("failed to get latest release: %w", err)
	}

	// mihomo amd64 特殊处理：按优先级逐个尝试候选版本
	if coreName == "mihomo" && runtime.GOARCH == "amd64" {
		return u.downloadMihomoWithFallback(coreInfo, release, coreDir, progressFn)
	}

	// 查找匹配当前平台的 asset
	downloadURL, err := u.findAssetURL(release.Assets, coreName)
	if err != nil {
		return fmt.Errorf("failed to find asset: %w", err)
	}

	// 构建候选下载 URL（镜像 + 原始）
	originalURL := ""
	for _, asset := range release.Assets {
		if strings.EqualFold(asset.Name, filepath.Base(downloadURL)) ||
			strings.Contains(asset.BrowserDownloadURL, filepath.Base(downloadURL)) {
			originalURL = asset.BrowserDownloadURL
			break
		}
	}

	// 尝试下载，支持降级
	tmpFile, err := os.CreateTemp("", "v2rayn-core-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	downloadErr := u.downloadFile(downloadURL, tmpPath, progressFn)
	if downloadErr != nil && originalURL != "" && originalURL != downloadURL {
		log.Printf("Mirror download failed for %s: %v, trying original URL...", coreName, downloadErr)
		// 清空临时文件重新下载
		os.Remove(tmpPath)
		downloadErr = u.downloadFile(originalURL, tmpPath, progressFn)
	}
	if downloadErr != nil {
		return fmt.Errorf("failed to download: %w", downloadErr)
	}

	// 解压到内核子目录
	binPath := filepath.Join(coreDir, coreInfo.BinaryName)
	if err := u.ExtractBinary(tmpPath, downloadURL, binPath, coreInfo.BinaryName); err != nil {
		return fmt.Errorf("failed to extract: %w", err)
	}

	// Linux/macOS 添加执行权限
	if runtime.GOOS != "windows" {
		if err := os.Chmod(binPath, 0755); err != nil {
			return fmt.Errorf("failed to set executable permission: %w", err)
		}
	}

	log.Printf("Successfully downloaded %s %s to %s", coreName, release.TagName, binPath)
	return nil
}

// downloadMihomoWithFallback mihomo amd64 专用下载逻辑：按 CPU 级别优先级逐个尝试候选版本
func (u *Updater) downloadMihomoWithFallback(coreInfo *CoreInfo, release *GitHubRelease, coreDir string, progressFn func(downloaded, total int64)) error {
	candidates, err := u.findMihomoAssets(release.Assets, runtime.GOOS)
	if err != nil {
		return err
	}

	binPath := filepath.Join(coreDir, coreInfo.BinaryName)
	var lastErr error

	for i, cand := range candidates {
		mirrorURL := u.getDownloadBaseURL(cand.url)

		// 找到原始 URL 用于镜像降级
		originalURL := cand.url

		log.Printf("Trying mihomo candidate %d/%d: %s", i+1, len(candidates), cand.name)

		// 创建临时文件
		tmpFile, err := os.CreateTemp("", "v2rayn-core-*.tmp")
		if err != nil {
			lastErr = fmt.Errorf("failed to create temp file: %w", err)
			continue
		}
		tmpPath := tmpFile.Name()
		tmpFile.Close()

		// 尝试下载（先镜像，后原始）
		downloadErr := u.downloadFile(mirrorURL, tmpPath, progressFn)
		if downloadErr != nil && mirrorURL != originalURL {
			log.Printf("Mirror download failed for %s: %v, trying original URL...", cand.name, downloadErr)
			os.Remove(tmpPath)
			tmpFile2, err := os.CreateTemp("", "v2rayn-core-*.tmp")
			if err != nil {
				lastErr = fmt.Errorf("failed to create temp file: %w", err)
				continue
			}
			tmpPath = tmpFile2.Name()
			tmpFile2.Close()
			downloadErr = u.downloadFile(originalURL, tmpPath, progressFn)
		}
		if downloadErr != nil {
			os.Remove(tmpPath)
			lastErr = fmt.Errorf("failed to download %s: %w", cand.name, downloadErr)
			log.Printf("Download failed for %s: %v, trying next candidate...", cand.name, downloadErr)
			continue
		}

		// 尝试解压
		extractErr := u.ExtractBinary(tmpPath, mirrorURL, binPath, coreInfo.BinaryName)
		os.Remove(tmpPath)
		if extractErr != nil {
			lastErr = fmt.Errorf("failed to extract from %s: %w", cand.name, extractErr)
			log.Printf("Extract failed for %s: %v, trying next candidate...", cand.name, extractErr)
			continue
		}

		// Linux/macOS 添加执行权限
		if runtime.GOOS != "windows" {
			if err := os.Chmod(binPath, 0755); err != nil {
				return fmt.Errorf("failed to set executable permission: %w", err)
			}
		}

		log.Printf("Successfully downloaded mihomo from %s to %s", cand.name, binPath)
		return nil
	}

	return fmt.Errorf("all mihomo candidates failed: %w", lastErr)
}

// DownloadCoreFromURL 从自定义 URL 下载内核
func (u *Updater) DownloadCoreFromURL(coreName, downloadURL string, progressFn func(downloaded, total int64)) error {
	cores := u.GetSupportedCores()
	var coreInfo *CoreInfo
	for i := range cores {
		if cores[i].Name == coreName {
			coreInfo = &cores[i]
			break
		}
	}
	if coreInfo == nil {
		return fmt.Errorf("unsupported core: %s", coreName)
	}

	// 确保内核子目录存在
	coreDir := filepath.Join(u.cfg.BinDir, coreInfo.SubDir)
	if err := os.MkdirAll(coreDir, 0755); err != nil {
		return fmt.Errorf("failed to create core directory: %w", err)
	}

	log.Printf("Downloading %s from custom URL: %s", coreName, downloadURL)

	// 下载到临时文件
	tmpFile, err := os.CreateTemp("", "v2rayn-core-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	if err := u.downloadFile(downloadURL, tmpPath, progressFn); err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}

	// 解压到内核子目录
	binPath := filepath.Join(coreDir, coreInfo.BinaryName)
	if err := u.ExtractBinary(tmpPath, downloadURL, binPath, coreInfo.BinaryName); err != nil {
		return fmt.Errorf("failed to extract: %w", err)
	}

	// Linux/macOS 添加执行权限
	if runtime.GOOS != "windows" {
		if err := os.Chmod(binPath, 0755); err != nil {
			return fmt.Errorf("failed to set executable permission: %w", err)
		}
	}

	log.Printf("Successfully downloaded %s from custom URL to %s", coreName, binPath)
	return nil
}

// getBaseURL 获取 GitHub API 基础 URL（支持镜像）
func (u *Updater) getBaseURL() string {
	if u.cfg.GitHubMirror != "" {
		mirror := strings.TrimRight(u.cfg.GitHubMirror, "/")
		// 如果镜像 URL 不包含 /api/ 路径，添加它
		if !strings.Contains(mirror, "/api/") {
			return mirror + "/api"
		}
		return mirror
	}
	return "https://api.github.com"
}

// getDownloadBaseURL 获取下载基础 URL（支持镜像加速）
func (u *Updater) getDownloadBaseURL(originalURL string) string {
	if u.cfg.GitHubMirror != "" {
		mirror := strings.TrimRight(u.cfg.GitHubMirror, "/")
		// 替换 github.com 为镜像地址
		return strings.Replace(originalURL, "https://github.com", mirror, 1)
	}
	return originalURL
}

// getLatestRelease 获取 GitHub 仓库的最新 release（支持镜像降级 + 缓存）
func (u *Updater) getLatestRelease(repo string) (*GitHubRelease, error) {
	// 先检查缓存
	u.cacheMu.RLock()
	if entry, ok := u.releaseCache[repo]; ok {
		if time.Since(entry.fetchedAt) < releaseCacheTTL {
			u.cacheMu.RUnlock()
			log.Printf("Cache hit for %s (age: %s)", repo, time.Since(entry.fetchedAt).Round(time.Second))
			return entry.release, nil
		}
	}
	u.cacheMu.RUnlock()

	// 构建候选 URL 列表：优先镜像，降级原站
	var candidateURLs []string
	mirrorBase := u.getBaseURL()
	originalBase := "https://api.github.com"

	primaryURL := fmt.Sprintf("%s/repos/%s/releases/latest", mirrorBase, repo)
	candidateURLs = append(candidateURLs, primaryURL)

	// 如果镜像不同于原站，添加原站作为降级
	originalURL := fmt.Sprintf("%s/repos/%s/releases/latest", originalBase, repo)
	if primaryURL != originalURL {
		candidateURLs = append(candidateURLs, originalURL)
	}

	var lastErr error
	for i, url := range candidateURLs {
		release, err := u.fetchRelease(url)
		if err == nil {
			if i > 0 {
				log.Printf("GitHub mirror failed, fallback to original succeeded for %s", repo)
			}
			// 写入缓存
			u.cacheMu.Lock()
			u.releaseCache[repo] = &releaseCacheEntry{
				release:   release,
				fetchedAt: time.Now(),
			}
			u.cacheMu.Unlock()
			return release, nil
		}
		lastErr = err
		if i == 0 && len(candidateURLs) > 1 {
			log.Printf("GitHub mirror request failed for %s: %v, trying original...", repo, err)
		}
	}

	return nil, fmt.Errorf("all GitHub API endpoints failed for %s: %w", repo, lastErr)
}

// fetchRelease 从指定 URL 获取 release 信息
func (u *Updater) fetchRelease(url string) (*GitHubRelease, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "v2rayN-Go/1.0")

	resp, err := u.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// detectX86Level 检测当前 CPU 支持的 x86-64 微架构级别
// 返回 1, 2, 3, 4 分别对应 x86-64-v1, v2, v3, v4
func detectX86Level() int {
	// v4: AVX512F
	if cpu.X86.HasAVX512F {
		return 4
	}
	// v3: AVX2 + FMA + BMI2 + ...
	if cpu.X86.HasAVX2 && cpu.X86.HasFMA && cpu.X86.HasBMI2 {
		return 3
	}
	// v2: SSE4.2 + POPCNT + CMPXCHG16B
	if cpu.X86.HasSSE42 && cpu.X86.HasPOPCNT && cpu.X86.HasCX16 {
		return 2
	}
	// v1: baseline (所有 x86-64 CPU)
	return 1
}

// hasGoVersionSuffix 检查文件名是否包含 Go 版本后缀（如 go120, go123, go124）
func hasGoVersionSuffix(name string) bool {
	// 匹配 -go120, -go121, -go122, -go123, -go124, -go125 等模式
	for i := 0; i < len(name)-4; i++ {
		if name[i] == '-' && name[i+1] == 'g' && name[i+2] == 'o' {
			// 检查后面是否是数字
			j := i + 3
			digitCount := 0
			for j < len(name) && name[j] >= '0' && name[j] <= '9' {
				digitCount++
				j++
			}
			if digitCount >= 2 {
				return true
			}
		}
	}
	return false
}

// findAssetURL 根据当前平台查找匹配的下载链接
// 每个内核使用不同的命名约定，需要分别处理
func (u *Updater) findAssetURL(assets []GitHubAsset, coreName string) (string, error) {
	osName := runtime.GOOS     // windows, linux, darwin
	archName := runtime.GOARCH // amd64, arm64, 386

	// 根据内核类型定义平台关键词映射
	type platformKeywords struct {
		osNames   []string // 可能的 OS 名称
		archNames []string // 可能的架构名称
	}

	var keywords platformKeywords

	switch coreName {
	case "xray":
		// Xray 命名: Xray-windows-64.zip, Xray-linux-arm64-v8a.zip
		keywords.osNames = []string{osName}
		switch archName {
		case "amd64":
			keywords.archNames = []string{"64"}
		case "arm64":
			keywords.archNames = []string{"arm64", "arm64-v8a"}
		case "386":
			keywords.archNames = []string{"32"}
		default:
			keywords.archNames = []string{archName}
		}

	case "sing-box":
		// Sing-box 命名: sing-box-1.x.x-windows-amd64.zip (Go 风格)
		keywords.osNames = []string{osName}
		keywords.archNames = []string{archName}

	case "mihomo":
		// Mihomo 命名复杂，有多个 amd64 变体（v1/v2/v3/compatible）
		// 需要特殊处理，返回第一个候选（DownloadCore 会处理降级重试）
		if archName == "amd64" {
			candidates, err := u.findMihomoAssets(assets, osName)
			if err != nil {
				return "", err
			}
			return u.getDownloadBaseURL(candidates[0].url), nil
		}
		keywords.osNames = []string{osName}
		keywords.archNames = []string{archName}

	default:
		keywords.osNames = []string{osName}
		keywords.archNames = []string{archName}
	}

	for _, asset := range assets {
		name := strings.ToLower(asset.Name)

		// 跳过非压缩包
		if !strings.HasSuffix(name, ".zip") && !strings.HasSuffix(name, ".tar.gz") && !strings.HasSuffix(name, ".tgz") {
			continue
		}

		// 检查 OS 匹配
		osMatch := false
		for _, osKey := range keywords.osNames {
			if strings.Contains(name, osKey) {
				osMatch = true
				break
			}
		}
		if !osMatch {
			continue
		}

		// 检查架构匹配
		archMatch := false
		for _, archKey := range keywords.archNames {
			if strings.Contains(name, archKey) {
				archMatch = true
				break
			}
		}
		if !archMatch {
			continue
		}

		return u.getDownloadBaseURL(asset.BrowserDownloadURL), nil
	}

	return "", fmt.Errorf("no matching asset found for %s/%s (core: %s)", osName, archName, coreName)
}

// mihomoCandidate mihomo 下载候选
type mihomoCandidate struct {
	name     string
	url      string
	priority int
}

// findMihomoAssets 专门为 mihomo amd64 查找所有匹配的 asset，按优先级排序返回
// mihomo 有大量 amd64 变体：compatible, v1, v2, v3, 默认，以及带 go 版本后缀的
func (u *Updater) findMihomoAssets(assets []GitHubAsset, osName string) ([]mihomoCandidate, error) {
	level := detectX86Level()
	log.Printf("Detected x86-64 level: v%d", level)

	var candidates []mihomoCandidate

	for _, asset := range assets {
		name := strings.ToLower(asset.Name)

		// 只处理压缩包
		if !strings.HasSuffix(name, ".zip") && !strings.HasSuffix(name, ".tar.gz") && !strings.HasSuffix(name, ".tgz") {
			continue
		}

		// 必须包含 OS 名称
		if !strings.Contains(name, osName) {
			continue
		}

		// 必须包含 amd64
		if !strings.Contains(name, "amd64") {
			continue
		}

		// 跳过带 go 版本后缀的文件
		if hasGoVersionSuffix(name) {
			continue
		}

		// 检查是否包含 v1/v2/v3 级别标记
		// 文件名格式: mihomo-windows-amd64-v3-v1.19.24.zip
		// 注意不要和版本号 v1.19.24 混淆
		hasV3 := strings.Contains(name, "-amd64-v3-") || strings.HasSuffix(name, "-amd64-v3.zip") || strings.HasSuffix(name, "-amd64-v3.tar.gz")
		hasV2 := strings.Contains(name, "-amd64-v2-") || strings.HasSuffix(name, "-amd64-v2.zip") || strings.HasSuffix(name, "-amd64-v2.tar.gz")
		hasV1 := strings.Contains(name, "-amd64-v1-") || strings.HasSuffix(name, "-amd64-v1.zip") || strings.HasSuffix(name, "-amd64-v1.tar.gz")
		hasCompatible := strings.Contains(name, "-amd64-compatible-") || strings.HasSuffix(name, "-amd64-compatible.zip") || strings.HasSuffix(name, "-amd64-compatible.tar.gz")

		// 根据 CPU 级别和文件变体计算优先级
		// 优先级越小越优先，优先匹配当前 CPU 级别对应的版本
		priority := 100

		switch level {
		case 4, 3:
			// v4/v3 CPU: 优先 v3，然后 v2，然后 v1，然后 default，最后 compatible
			switch {
			case hasV3:
				priority = 1
			case hasV2:
				priority = 2
			case hasV1:
				priority = 3
			default:
				if hasCompatible {
					priority = 5
				} else {
					priority = 4
				}
			}
		case 2:
			// v2 CPU: 优先 v2，然后 v1，然后 default，最后 compatible
			switch {
			case hasV2:
				priority = 1
			case hasV1:
				priority = 2
			default:
				if hasCompatible {
					priority = 4
				} else {
					priority = 3
				}
			}
		default:
			// v1 CPU: 优先 v1，然后 compatible，然后 default
			switch {
			case hasV1:
				priority = 1
			default:
				if hasCompatible {
					priority = 2
				} else {
					priority = 3
				}
			}
		}

		candidates = append(candidates, mihomoCandidate{
			name:     asset.Name,
			url:      asset.BrowserDownloadURL,
			priority: priority,
		})
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no matching mihomo asset found for %s/amd64", osName)
	}

	// 按优先级排序（升序）
	for i := 0; i < len(candidates)-1; i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].priority < candidates[i].priority {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	for i, c := range candidates {
		log.Printf("Mihomo candidate %d: %s (priority %d)", i, c.name, c.priority)
	}

	return candidates, nil
}

// ExtractBinary 从压缩包中提取可执行文件到目标路径（公开方法，供外部调用）
func (u *Updater) ExtractBinary(archivePath, downloadURL, destPath, binaryName string) error {
	if strings.HasSuffix(downloadURL, ".zip") || strings.HasSuffix(strings.ToLower(downloadURL), ".zip") {
		return extractFromZip(archivePath, destPath, binaryName)
	}
	if strings.HasSuffix(downloadURL, ".tar.gz") || strings.HasSuffix(downloadURL, ".tgz") {
		return extractFromTarGz(archivePath, destPath, binaryName)
	}

	// 如果不是压缩包，直接复制
	return copyFile(archivePath, destPath)
}

// extractFromZip 从 zip 文件中提取可执行文件
func extractFromZip(zipPath, destPath, binaryName string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	// 第一步：精确匹配目标二进制名
	for _, f := range r.File {
		baseName := filepath.Base(f.Name)
		if matchBinaryName(baseName, binaryName) {
			return extractZipFile(f, destPath)
		}
	}

	// 第二步：模糊匹配 — 文件名包含目标名（不含扩展名）的可执行文件
	// 例如 mihomo-windows-amd64-v3.exe 包含 "mihomo"
	targetClean := strings.TrimSuffix(strings.ToLower(binaryName), ".exe")
	for _, f := range r.File {
		baseName := strings.ToLower(filepath.Base(f.Name))
		if isExecutable(baseName) && strings.Contains(baseName, targetClean) {
			log.Printf("Fuzzy match: extracting %s as %s", f.Name, binaryName)
			return extractZipFile(f, destPath)
		}
	}

	// 第三步：如果压缩包中只有一个可执行文件，直接提取并重命名
	var executables []*zip.File
	for _, f := range r.File {
		baseName := filepath.Base(f.Name)
		if isExecutable(baseName) {
			executables = append(executables, f)
		}
	}
	if len(executables) == 1 {
		log.Printf("Single executable found: extracting %s as %s", executables[0].Name, binaryName)
		return extractZipFile(executables[0], destPath)
	}

	return fmt.Errorf("binary %s not found in zip archive", binaryName)
}

// extractZipFile 从 zip 中提取单个文件到目标路径
func extractZipFile(f *zip.File, destPath string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	outFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, rc)
	return err
}

// isExecutable 检查文件名是否是可执行文件
func isExecutable(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".exe") || (!strings.Contains(filepath.Base(lower), ".") && !strings.HasSuffix(lower, "/"))
}

// extractFromTarGz 从 tar.gz 文件中提取可执行文件
func extractFromTarGz(tarPath, destPath, binaryName string) error {
	f, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	// 收集所有可执行文件信息（用于回退匹配）
	type tarEntry struct {
		header *tar.Header
		name   string
	}

	tr := tar.NewReader(gz)
	targetClean := strings.TrimSuffix(strings.ToLower(binaryName), ".exe")
	var executables []tarEntry

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		baseName := filepath.Base(header.Name)

		// 精确匹配
		if matchBinaryName(baseName, binaryName) {
			return extractTarEntry(tr, destPath)
		}

		// 收集可执行文件用于回退
		if isExecutable(baseName) {
			executables = append(executables, tarEntry{header: header, name: baseName})
			// 模糊匹配
			if strings.Contains(strings.ToLower(baseName), targetClean) {
				log.Printf("Fuzzy match: extracting %s as %s", header.Name, binaryName)
				return extractTarEntry(tr, destPath)
			}
		}
	}

	// 如果只有一个可执行文件，提取并重命名
	if len(executables) == 1 {
		// 需要重新打开文件并定位到该条目
		f2, err := os.Open(tarPath)
		if err != nil {
			return err
		}
		defer f2.Close()

		gz2, err := gzip.NewReader(f2)
		if err != nil {
			return err
		}
		defer gz2.Close()

		tr2 := tar.NewReader(gz2)
		for {
			header, err := tr2.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			if header.Typeflag == tar.TypeReg && filepath.Base(header.Name) == executables[0].name {
				log.Printf("Single executable found: extracting %s as %s", header.Name, binaryName)
				return extractTarEntry(tr2, destPath)
			}
		}
	}

	return fmt.Errorf("binary %s not found in tar.gz archive", binaryName)
}

// extractTarEntry 从 tar 流中提取当前条目到目标路径
func extractTarEntry(tr *tar.Reader, destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	outFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, tr)
	return err
}

// matchBinaryName 检查文件名是否匹配目标二进制名
func matchBinaryName(fileName, targetName string) bool {
	fileNameLower := strings.ToLower(fileName)
	targetLower := strings.ToLower(targetName)

	// 精确匹配
	if fileNameLower == targetLower {
		return true
	}

	// 去掉 .exe 后缀匹配
	fileNameClean := strings.TrimSuffix(fileNameLower, ".exe")
	targetClean := strings.TrimSuffix(targetLower, ".exe")
	if fileNameClean == targetClean {
		return true
	}

	return false
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// downloadFile 下载文件到指定路径
func (u *Updater) downloadFile(url string, destPath string, progressFn func(downloaded, total int64)) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "v2rayN-Go/1.0")

	resp, err := u.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	totalSize := resp.ContentLength
	var downloaded int64

	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := file.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
			downloaded += int64(n)
			if progressFn != nil {
				progressFn(downloaded, totalSize)
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return readErr
		}
	}

	return nil
}

// getBinaryName 根据平台返回正确的可执行文件名
func getBinaryName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}
