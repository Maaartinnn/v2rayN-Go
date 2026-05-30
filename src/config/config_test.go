package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"v2rayn-go/coredef"
)

// ========== AtomicWriteFile 测试 ==========

func TestAtomicWriteFile_BasicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.json")

	data := []byte(`{"key": "value"}`)
	if err := AtomicWriteFile(filePath, data, 0644); err != nil {
		t.Fatalf("AtomicWriteFile failed: %v", err)
	}

	// 验证文件内容正确
	got, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	if string(got) != string(data) {
		t.Fatalf("file content mismatch: got %q, want %q", string(got), string(data))
	}

	// 验证没有残留的临时文件
	entries, _ := os.ReadDir(tmpDir)
	for _, e := range entries {
		if len(e.Name()) > 0 && e.Name()[:4] == ".tmp" {
			t.Fatalf("temporary file left behind: %s", e.Name())
		}
	}
}

func TestAtomicWriteFile_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "overwrite.json")

	// 先写入初始内容
	if err := AtomicWriteFile(filePath, []byte("old content"), 0644); err != nil {
		t.Fatalf("first write failed: %v", err)
	}

	// 覆盖写入新内容
	newData := []byte(`{"updated": true}`)
	if err := AtomicWriteFile(filePath, newData, 0644); err != nil {
		t.Fatalf("overwrite failed: %v", err)
	}

	got, _ := os.ReadFile(filePath)
	if string(got) != string(newData) {
		t.Fatalf("overwrite content mismatch: got %q, want %q", string(got), string(newData))
	}
}

func TestAtomicWriteFile_Perm(t *testing.T) {
	// Windows 不支持 Unix 风格的文件权限位，跳过此测试
	if runtime.GOOS == "windows" {
		t.Skip("skipping permission test on Windows (chmod is no-op)")
	}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "perm.json")

	if err := AtomicWriteFile(filePath, []byte("data"), 0644); err != nil {
		t.Fatalf("AtomicWriteFile failed: %v", err)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}

	// 检查权限（仅检查低 12 位，忽略 setuid/setgid/sticky）
	perm := info.Mode().Perm()
	if perm != 0644 {
		t.Fatalf("file permission = %o, want 0644", perm)
	}
}

// ========== loadJSONConfig .bak 容灾回滚测试 ==========

func TestLoadJSONConfig_0KB_RestoreFromBackup(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &AppConfig{AppDir: tmpDir}
	cfg.ListenIP = "127.0.0.1"
	cfg.WebPort = 9999

	configPath := filepath.Join(tmpDir, "config.json")
	bakPath := configPath + ".bak"

	// 1. 先写一份正常的 .bak 备份
	bakData := []byte(`{"web_port": 7777, "listen_ip": "0.0.0.0"}`)
	if err := AtomicWriteFile(bakPath, bakData, 0644); err != nil {
		t.Fatalf("failed to write .bak: %v", err)
	}

	// 2. 创建一个 0KB 的 config.json（模拟断电导致的空文件）
	if err := os.WriteFile(configPath, []byte{}, 0644); err != nil {
		t.Fatalf("failed to write 0KB config: %v", err)
	}

	// 3. 调用 loadJSONConfig，期望从 .bak 恢复
	cfg.loadJSONConfig()

	if cfg.WebPort != 7777 {
		t.Fatalf("expected WebPort 7777 (restored from .bak), got %d", cfg.WebPort)
	}
	if cfg.ListenIP != "0.0.0.0" {
		t.Fatalf("expected ListenIP '0.0.0.0' (restored from .bak), got %q", cfg.ListenIP)
	}

	// 4. 验证 config.json 已被恢复内容覆盖（原子写入）
	restored, _ := os.ReadFile(configPath)
	if string(restored) != string(bakData) {
		t.Fatalf("config.json was not restored: got %q", string(restored))
	}
}

func TestLoadJSONConfig_CorruptJSON_RestoreFromBackup(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &AppConfig{AppDir: tmpDir}
	cfg.ListenIP = "127.0.0.1"
	cfg.WebPort = 9999

	configPath := filepath.Join(tmpDir, "config.json")
	bakPath := configPath + ".bak"

	// 1. 写入正常的 .bak
	bakData := []byte(`{"web_port": 8888, "listen_ip": "10.0.0.1"}`)
	if err := AtomicWriteFile(bakPath, bakData, 0644); err != nil {
		t.Fatalf("failed to write .bak: %v", err)
	}

	// 2. 写入损坏的 config.json（模拟用户手动编辑出错或断电残缺）
	if err := os.WriteFile(configPath, []byte(`{"web_port": 8080, "listen_ip": `), 0644); err != nil {
		t.Fatalf("failed to write corrupt config: %v", err)
	}

	// 3. 调用 loadJSONConfig，期望从 .bak 恢复
	cfg.loadJSONConfig()

	if cfg.WebPort != 8888 {
		t.Fatalf("expected WebPort 8888 (restored from .bak), got %d", cfg.WebPort)
	}
	if cfg.ListenIP != "10.0.0.1" {
		t.Fatalf("expected ListenIP '10.0.0.1' (restored from .bak), got %q", cfg.ListenIP)
	}
}

func TestLoadJSONConfig_NoBackup_FallbackToDefaults(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := DefaultConfig()
	cfg.AppDir = tmpDir
	originalPort := cfg.WebPort

	configPath := filepath.Join(tmpDir, "config.json")

	// 1. 写入损坏的 config.json，但没有 .bak 文件
	if err := os.WriteFile(configPath, []byte(`{broken json`), 0644); err != nil {
		t.Fatalf("failed to write corrupt config: %v", err)
	}

	// 2. 调用 loadJSONConfig，期望回退到默认值（即 DefaultConfig 的值不变）
	cfg.loadJSONConfig()

	if cfg.WebPort != originalPort {
		t.Fatalf("expected WebPort %d (default), got %d", originalPort, cfg.WebPort)
	}
}

func TestLoadJSONConfig_NormalLoad(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &AppConfig{AppDir: tmpDir}
	configPath := filepath.Join(tmpDir, "config.json")

	// 写入正常的 config.json
	if err := os.WriteFile(configPath, []byte(`{"web_port": 5555, "listen_ip": "192.168.1.1"}`), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg.loadJSONConfig()

	if cfg.WebPort != 5555 {
		t.Fatalf("expected WebPort 5555, got %d", cfg.WebPort)
	}
	if cfg.ListenIP != "192.168.1.1" {
		t.Fatalf("expected ListenIP '192.168.1.1', got %q", cfg.ListenIP)
	}
}

// ========== BackupConfig 测试 ==========

func TestBackupConfig_CreatesBackup(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &AppConfig{
		AppDir:   tmpDir,
		WebPort:  1234,
		ListenIP: "127.0.0.1",
	}

	// 先保存一份 config.json
	if err := cfg.SaveJSONConfig(); err != nil {
		t.Fatalf("SaveJSONConfig failed: %v", err)
	}

	// 调用 BackupConfig
	cfg.BackupConfig()

	bakPath := filepath.Join(tmpDir, "config.json.bak")
	bakData, err := os.ReadFile(bakPath)
	if err != nil {
		t.Fatalf("BackupConfig did not create .bak file: %v", err)
	}

	// 验证 .bak 内容与 config.json 一致
	configData, _ := os.ReadFile(filepath.Join(tmpDir, "config.json"))
	if string(bakData) != string(configData) {
		t.Fatalf(".bak content mismatch with config.json")
	}
}

func TestBackupConfig_OnlyOnce(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &AppConfig{
		AppDir:   tmpDir,
		WebPort:  1111,
		ListenIP: "127.0.0.1",
	}

	if err := cfg.SaveJSONConfig(); err != nil {
		t.Fatalf("SaveJSONConfig failed: %v", err)
	}

	// 第一次备份
	cfg.BackupConfig()
	bakPath := filepath.Join(tmpDir, "config.json.bak")
	firstBak, _ := os.ReadFile(bakPath)

	// 修改配置并保存
	cfg.WebPort = 2222
	cfg.SaveJSONConfig()

	// 再次调用 BackupConfig（sync.Once 保证不会重复执行）
	cfg.BackupConfig()
	secondBak, _ := os.ReadFile(bakPath)

	// .bak 内容应该还是第一次备份的内容（1111）
	if string(firstBak) != string(secondBak) {
		t.Fatalf("BackupConfig executed more than once (sync.Once failed)")
	}

	// 验证 .bak 中的端口是第一次的值
	var bakCfg AppConfig
	json.Unmarshal(secondBak, &bakCfg)
	if bakCfg.WebPort != 1111 {
		t.Fatalf("expected .bak WebPort 1111, got %d", bakCfg.WebPort)
	}
}

func TestBackupConfig_OnlyWritesWhenConfigValid(t *testing.T) {
	tmpDir := t.TempDir()

	// 没有 config.json 的情况
	cfg := &AppConfig{AppDir: tmpDir}
	cfg.BackupConfig()

	bakPath := filepath.Join(tmpDir, "config.json.bak")
	if _, err := os.Stat(bakPath); !os.IsNotExist(err) {
		t.Fatal("BackupConfig should not create .bak when config.json does not exist")
	}
}

// ========== SaveJSONConfig 原子写入测试 ==========

func TestSaveJSONConfig_AtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &AppConfig{
		AppDir:   tmpDir,
		WebPort:  3333,
		ListenIP: "10.0.0.1",
	}

	if err := cfg.SaveJSONConfig(); err != nil {
		t.Fatalf("SaveJSONConfig failed: %v", err)
	}

	// 验证文件内容
	configPath := filepath.Join(tmpDir, "config.json")
	data, _ := os.ReadFile(configPath)

	var loaded AppConfig
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("failed to parse saved config: %v", err)
	}

	if loaded.WebPort != 3333 {
		t.Fatalf("expected WebPort 3333, got %d", loaded.WebPort)
	}
	if loaded.ListenIP != "10.0.0.1" {
		t.Fatalf("expected ListenIP '10.0.0.1', got %q", loaded.ListenIP)
	}

	// 验证没有残留临时文件
	entries, _ := os.ReadDir(tmpDir)
	for _, e := range entries {
		if len(e.Name()) > 4 && e.Name()[:4] == ".tmp" {
			t.Fatalf("temporary file left behind: %s", e.Name())
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.WebPort != coredef.DefaultWebPort {
		t.Fatalf("expected WebPort %d, got %d", coredef.DefaultWebPort, cfg.WebPort)
	}
	if cfg.ListenIP != coredef.DefaultListenIP {
		t.Fatalf("expected ListenIP '%s', got '%s'", coredef.DefaultListenIP, cfg.ListenIP)
	}
	if cfg.SocksPort != coredef.DefaultSocksPort {
		t.Fatalf("expected SocksPort %d, got %d", coredef.DefaultSocksPort, cfg.SocksPort)
	}
	if cfg.HTTPPort != coredef.DefaultHTTPPort {
		t.Fatalf("expected HTTPPort %d, got %d", coredef.DefaultHTTPPort, cfg.HTTPPort)
	}
	if cfg.OutboundIP != coredef.DefaultOutboundIP {
		t.Fatalf("expected OutboundIP '%s', got '%s'", coredef.DefaultOutboundIP, cfg.OutboundIP)
	}
}

func TestGetListenAddr(t *testing.T) {
	cfg := &AppConfig{ListenIP: "0.0.0.0", WebPort: 8080}
	addr := cfg.GetListenAddr()
	if addr != "0.0.0.0:8080" {
		t.Fatalf("expected '0.0.0.0:8080', got '%s'", addr)
	}
}

func TestSaveAndLoadJSONConfig(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	cfg := &AppConfig{
		AppDir:     tmpDir,
		ListenIP:   "0.0.0.0",
		WebPort:    9999,
		SocksPort:  20808,
		HTTPPort:   20809,
		OutboundIP: "10.0.0.1",
	}

	// Save
	if err := cfg.SaveJSONConfig(); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Verify file exists
	configPath := filepath.Join(tmpDir, "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	// Parse and verify
	var loaded AppConfig
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	if loaded.WebPort != 9999 {
		t.Fatalf("expected WebPort 9999, got %d", loaded.WebPort)
	}
	if loaded.ListenIP != "0.0.0.0" {
		t.Fatalf("expected ListenIP '0.0.0.0', got '%s'", loaded.ListenIP)
	}
	if loaded.SocksPort != 20808 {
		t.Fatalf("expected SocksPort 20808, got %d", loaded.SocksPort)
	}
	if loaded.HTTPPort != 20809 {
		t.Fatalf("expected HTTPPort 20809, got %d", loaded.HTTPPort)
	}
	if loaded.OutboundIP != "10.0.0.1" {
		t.Fatalf("expected OutboundIP '10.0.0.1', got '%s'", loaded.OutboundIP)
	}
}

func TestSaveJSONConfig_OmitsInternalFields(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &AppConfig{
		AppDir:   tmpDir,
		DBPath:   "/some/path/db",
		BinDir:   "/some/path/bin",
		LogDir:   "/some/path/log",
		WebPort:  8080,
		ListenIP: "127.0.0.1",
	}

	if err := cfg.SaveJSONConfig(); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(tmpDir, "config.json"))
	jsonStr := string(data)

	// Internal fields should NOT appear in JSON
	if contains(jsonStr, "app_dir") || contains(jsonStr, "db_path") || contains(jsonStr, "bin_dir") || contains(jsonStr, "log_dir") {
		t.Fatal("internal fields should not appear in config.json")
	}
}

func TestSaveJSONConfig_OmitZeroGitHubMirror(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &AppConfig{
		AppDir:       tmpDir,
		GitHubMirror: "", // zero value
		WebPort:      8080,
	}

	if err := cfg.SaveJSONConfig(); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(tmpDir, "config.json"))
	jsonStr := string(data)

	if contains(jsonStr, "github_mirror") {
		t.Fatal("empty github_mirror should be omitted due to omitzero")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
