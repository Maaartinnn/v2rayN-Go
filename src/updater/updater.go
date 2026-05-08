package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
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

	// 确保内核子目录存在
	coreDir := filepath.Join(u.cfg.BinDir, coreInfo.SubDir)
	if err := os.MkdirAll(coreDir, 0755); err != nil {
		return fmt.Errorf("failed to create core directory: %w", err)
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
	if err := u.extractBinary(tmpPath, downloadURL, binPath, coreInfo.BinaryName); err != nil {
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
		// Mihomo 命名: mihomo-windows-amd64-v1.x.x.zip (Go 风格)
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

		return asset.BrowserDownloadURL, nil
	}

	return "", fmt.Errorf("no matching asset found for %s/%s (core: %s)", osName, archName, coreName)
}

// extractBinary 从压缩包中提取可执行文件到目标路径
func (u *Updater) extractBinary(archivePath, downloadURL, destPath, binaryName string) error {
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

	// 查找匹配的可执行文件
	for _, f := range r.File {
		baseName := filepath.Base(f.Name)
		// 匹配二进制文件名（忽略大小写和 .exe 后缀）
		if matchBinaryName(baseName, binaryName) {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			// 确保目标目录存在
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
	}

	return fmt.Errorf("binary %s not found in zip archive", binaryName)
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

	tr := tar.NewReader(gz)
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
		if matchBinaryName(baseName, binaryName) {
			// 确保目标目录存在
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
	}

	return fmt.Errorf("binary %s not found in tar.gz archive", binaryName)
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
