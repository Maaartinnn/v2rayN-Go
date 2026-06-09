package service

import (
	"testing"

	"v2rayn-go/config"
	"v2rayn-go/database"
)

// ========== SettingsService 测试 ==========

func TestGetSettings_IncludesCoreConfigDebug(t *testing.T) {
	database.InitTestDB()
	t.Cleanup(database.CleanupTestDB)

	cfg := &config.AppConfig{
		ListenIP:        "127.0.0.1",
		WebPort:         2017,
		SocksPort:       10808,
		HTTPPort:        10809,
		OutboundIP:      "0.0.0.0",
		GitHubMirror:    "https://mirror.example.com",
		CoreConfigDebug: true,
	}

	svc := NewSettingsService(cfg)
	settings := svc.GetSettings()

	if settings["core_config_debug"] != true {
		t.Fatalf("expected core_config_debug=true, got %v", settings["core_config_debug"])
	}
	if settings["listen_ip"] != "127.0.0.1" {
		t.Fatalf("expected listen_ip=127.0.0.1, got %v", settings["listen_ip"])
	}
	if settings["web_port"] != 2017 {
		t.Fatalf("expected web_port=2017, got %v", settings["web_port"])
	}
}

func TestGetSettings_CoreConfigDebug_DefaultFalse(t *testing.T) {
	database.InitTestDB()
	t.Cleanup(database.CleanupTestDB)

	cfg := config.DefaultConfig()
	cfg.AppDir = t.TempDir()

	svc := NewSettingsService(cfg)
	settings := svc.GetSettings()

	if settings["core_config_debug"] != false {
		t.Fatalf("expected core_config_debug=false (default), got %v", settings["core_config_debug"])
	}
}

func TestUpdateSettings_CoreConfigDebug_Enable(t *testing.T) {
	database.InitTestDB()
	t.Cleanup(database.CleanupTestDB)

	cfg := config.DefaultConfig()
	cfg.AppDir = t.TempDir()

	svc := NewSettingsService(cfg)

	// 默认是 false
	if cfg.CoreConfigDebug {
		t.Fatal("expected CoreConfigDebug=false before update")
	}

	// 开启调试模式
	enable := true
	err := svc.UpdateSettings(&UpdateSettingsRequest{
		CoreConfigDebug: &enable,
	})
	if err != nil {
		t.Fatalf("UpdateSettings failed: %v", err)
	}

	if !cfg.CoreConfigDebug {
		t.Fatal("expected CoreConfigDebug=true after update")
	}
}

func TestUpdateSettings_CoreConfigDebug_Disable(t *testing.T) {
	database.InitTestDB()
	t.Cleanup(database.CleanupTestDB)

	cfg := config.DefaultConfig()
	cfg.AppDir = t.TempDir()
	cfg.CoreConfigDebug = true

	svc := NewSettingsService(cfg)

	// 关闭调试模式
	disable := false
	err := svc.UpdateSettings(&UpdateSettingsRequest{
		CoreConfigDebug: &disable,
	})
	if err != nil {
		t.Fatalf("UpdateSettings failed: %v", err)
	}

	if cfg.CoreConfigDebug {
		t.Fatal("expected CoreConfigDebug=false after update")
	}
}

func TestUpdateSettings_CoreConfigDebug_NilIgnored(t *testing.T) {
	database.InitTestDB()
	t.Cleanup(database.CleanupTestDB)

	cfg := config.DefaultConfig()
	cfg.AppDir = t.TempDir()
	cfg.CoreConfigDebug = true

	svc := NewSettingsService(cfg)

	// 不传 CoreConfigDebug（nil），应保持原值
	err := svc.UpdateSettings(&UpdateSettingsRequest{
		ListenIP: strPtr("10.0.0.1"),
	})
	if err != nil {
		t.Fatalf("UpdateSettings failed: %v", err)
	}

	if !cfg.CoreConfigDebug {
		t.Fatal("expected CoreConfigDebug=true (unchanged when nil)")
	}
	if cfg.ListenIP != "10.0.0.1" {
		t.Fatalf("expected ListenIP=10.0.0.1, got %q", cfg.ListenIP)
	}
}

func strPtr(s string) *string {
	return &s
}
