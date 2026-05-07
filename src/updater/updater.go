package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"v2rayn-go/config"
)

// CoreInfo 内核信息
type CoreInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Repo        string `json:"repo"`           // GitHub 仓库，如 "XTLS/Xray-core"
	Version     string `json:"version"`        // 当前安装版本
	LatestVer   string `json:"latest_version"` // 最新可用版本
	BinaryName  string `json:"binary_name"`    // 可执行文件名
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

// Updater 内核更新管理器
type Updater struct {
	cfg    *config.AppConfig
	client *http.Client
}

// NewUpdater 创建更新管理器
func NewUpdater(cfg *config.AppConfig) *Updater {
	return &Updater{
		cfg: cfg,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
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
		},
		{
			Name:        "sing-box",
			DisplayName: "Sing-box",
			Repo:        "SagerNet/sing-box",
			BinaryName:  getBinaryName("sing-box"),
		},
	}
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

	// 检查当前安装版本
	binPath := filepath.Join(u.cfg.BinDir, coreInfo.BinaryName)
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
		binPath := filepath.Join(u.cfg.BinDir, cores[i].BinaryName)
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

// DownloadCore 下载指定内核的最新版本
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

	// 获取最新 release
	release, err := u.getLatestRelease(coreInfo.Repo)
	if err != nil {
		return fmt.Errorf("failed to get latest release: %w", err)
	}

	// 查找匹配当前平台的 asset
	downloadURL, err := u.findAssetURL(release.Assets, coreName)
	if err != nil {
		return fmt.Errorf("failed to find asset: %w", err)
	}

	log.Printf("Downloading %s from %s", coreName, downloadURL)

	// 下载文件
	binPath := filepath.Join(u.cfg.BinDir, coreInfo.BinaryName)
	if err := u.downloadFile(downloadURL, binPath, progressFn); err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}

	// Linux/macOS 添加执行权限
	if runtime.GOOS != "windows" {
		if err := os.Chmod(binPath, 0755); err != nil {
			return fmt.Errorf("failed to set executable permission: %w", err)
		}
	}

	log.Printf("Successfully downloaded %s %s", coreName, release.TagName)
	return nil
}

// getLatestRelease 获取 GitHub 仓库的最新 release
func (u *Updater) getLatestRelease(repo string) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

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
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// findAssetURL 根据当前平台查找匹配的下载链接
func (u *Updater) findAssetURL(assets []GitHubAsset, coreName string) (string, error) {
	osName := runtime.GOOS     // windows, linux, darwin
	archName := runtime.GOARCH // amd64, arm64

	// 映射架构名
	switch archName {
	case "amd64":
		archName = "amd64"
	case "arm64":
		archName = "arm64"
	case "386":
		archName = "386"
	}

	for _, asset := range assets {
		name := strings.ToLower(asset.Name)

		// 跳过非压缩包
		if !strings.HasSuffix(name, ".zip") && !strings.HasSuffix(name, ".tar.gz") && !strings.HasSuffix(name, ".tgz") {
			continue
		}

		// 检查是否匹配当前平台
		if strings.Contains(name, osName) && strings.Contains(name, archName) {
			return asset.BrowserDownloadURL, nil
		}
	}

	return "", fmt.Errorf("no matching asset found for %s/%s", osName, archName)
}

// downloadFile 下载文件到指定路径
func (u *Updater) downloadFile(url string, destPath string, progressFn func(downloaded, total int64)) error {
	// 先下载到临时文件
	tmpPath := destPath + ".tmp"
	defer os.Remove(tmpPath)

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

	file, err := os.Create(tmpPath)
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

	file.Close()

	// 替换目标文件
	if _, err := os.Stat(destPath); err == nil {
		os.Remove(destPath)
	}
	return os.Rename(tmpPath, destPath)
}

// getBinaryName 根据平台返回正确的可执行文件名
func getBinaryName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}
