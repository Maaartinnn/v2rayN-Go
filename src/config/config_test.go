package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"v2rayn-go/coredef"
)

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
