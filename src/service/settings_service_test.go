package service

import (
	"testing"

	"v2rayn-go/config"
	"v2rayn-go/database"
)

func setupSettingsTestDB(t *testing.T) {
	t.Helper()
	database.InitTestDB()
	t.Cleanup(database.CleanupTestDB)
}

func strPtr(s string) *string { return &s }

func TestGetSettingFast_CacheHit(t *testing.T) {
	setupSettingsTestDB(t)
	svc := NewSettingsService(&config.AppConfig{})
	svc.cacheMu.Lock()
	svc.cacheData["jwt_expire_hours"] = "48"
	svc.cacheMu.Unlock()
	got := svc.GetSettingFast("jwt_expire_hours")
	if got != "48" {
		t.Fatalf("expected '48', got %q", got)
	}
}

func TestGetSettingFast_CacheMiss_ThenDB(t *testing.T) {
	setupSettingsTestDB(t)
	svc := NewSettingsService(&config.AppConfig{})
	database.DB.Create(&database.AppSetting{Key: "jwt_expire_hours", Value: "24"})
	got := svc.GetSettingFast("jwt_expire_hours")
	if got != "24" {
		t.Fatalf("expected '24', got %q", got)
	}
	svc.cacheMu.RLock()
	cached, ok := svc.cacheData["jwt_expire_hours"]
	svc.cacheMu.RUnlock()
	if !ok || cached != "24" {
		t.Fatalf("expected cache to have '24', got %q (exists: %v)", cached, ok)
	}
}

func TestGetSettingFast_DCLDoubleCheck(t *testing.T) {
	setupSettingsTestDB(t)
	svc := NewSettingsService(&config.AppConfig{})
	database.DB.Create(&database.AppSetting{Key: "jwt_expire_hours", Value: "12"})
	for i := 0; i < 100; i++ {
		got := svc.GetSettingFast("jwt_expire_hours")
		if got != "12" {
			t.Fatalf("iteration %d: expected '12', got %q", i, got)
		}
	}
	svc.cacheMu.RLock()
	count := len(svc.cacheData)
	svc.cacheMu.RUnlock()
	if count != 1 {
		t.Fatalf("expected 1 cached key, got %d", count)
	}
}

func TestGetSettingFast_UpdateSettings(t *testing.T) {
	setupSettingsTestDB(t)
	svc := NewSettingsService(&config.AppConfig{})
	database.DB.Create(&database.AppSetting{Key: "jwt_expire_hours", Value: "24"})
	got := svc.GetSettingFast("jwt_expire_hours")
	if got != "24" {
		t.Fatalf("expected '24', got %q", got)
	}
	req := &UpdateSettingsRequest{JwtExpireHours: strPtr("720")}
	if err := svc.UpdateSettings(req); err != nil {
		t.Fatalf("UpdateSettings failed: %v", err)
	}
	got = svc.GetSettingFast("jwt_expire_hours")
	if got != "720" {
		t.Fatalf("expected '720', got %q", got)
	}
}
