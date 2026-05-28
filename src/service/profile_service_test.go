package service

import (
	"testing"
	"v2rayn-go/database"
)

func setupServiceTestDB(t *testing.T) {
	t.Helper()
	database.InitTestDB()
	t.Cleanup(database.CleanupTestDB)
}

func createTestGroup(t *testing.T) *database.NodeGroup {
	t.Helper()
	group := &database.NodeGroup{
		UUID:      database.GenerateUUID(),
		Alias:     "Test Group",
		SortOrder: 10,
		Enabled:   true,
	}
	if err := database.DB.Create(group).Error; err != nil {
		t.Fatalf("failed to create test group: %v", err)
	}
	return group
}

// ==================== ProfileService ====================

func TestProfileService_Create(t *testing.T) {
	setupServiceTestDB(t)
	group := createTestGroup(t)

	svc := NewProfileService()
	profile := &database.Profile{
		Name:          "TestNode",
		ProxyAddress:  "example.com",
		ProxyPort:     443,
		ProxyProtocol: "vless",
		GroupUUID:     group.UUID,
	}

	if err := svc.Create(profile); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.UUID == "" {
		t.Fatal("expected UUID to be set")
	}
	if profile.SortOrder != 10 {
		t.Fatalf("expected sort_order 10, got %d", profile.SortOrder)
	}
}

func TestProfileService_Create_NoGroupUUID(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewProfileService()
	profile := &database.Profile{Name: "Test"}
	err := svc.Create(profile)
	if err == nil {
		t.Fatal("expected error for missing group_uuid")
	}
	var validationErr *ErrValidation
	if !isErrValidation(err, validationErr) {
		t.Fatalf("expected ErrValidation, got %T", err)
	}
}

func TestProfileService_Create_InvalidGroup(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewProfileService()
	profile := &database.Profile{
		Name:      "Test",
		GroupUUID: "nonexistent-uuid",
	}
	err := svc.Create(profile)
	if err == nil {
		t.Fatal("expected error for nonexistent group")
	}
}

func TestProfileService_Get(t *testing.T) {
	setupServiceTestDB(t)
	group := createTestGroup(t)

	svc := NewProfileService()
	profile := &database.Profile{
		UUID:          database.GenerateUUID(),
		Name:          "GetMe",
		ProxyAddress:  "host.com",
		ProxyPort:     443,
		ProxyProtocol: "trojan",
		GroupUUID:     group.UUID,
		SortOrder:     10,
	}
	database.DB.Create(profile)

	found, err := svc.Get(profile.UUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.Name != "GetMe" {
		t.Fatalf("expected name 'GetMe', got '%s'", found.Name)
	}
}

func TestProfileService_Get_NotFound(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewProfileService()
	_, err := svc.Get("nonexistent-uuid")
	if err == nil {
		t.Fatal("expected error for nonexistent profile")
	}
}

func TestProfileService_List(t *testing.T) {
	setupServiceTestDB(t)
	group := createTestGroup(t)

	svc := NewProfileService()
	database.DB.Create(&database.Profile{UUID: database.GenerateUUID(), Name: "A", ProxyProtocol: "vless", GroupUUID: group.UUID, SortOrder: 10})
	database.DB.Create(&database.Profile{UUID: database.GenerateUUID(), Name: "B", ProxyProtocol: "trojan", GroupUUID: group.UUID, SortOrder: 20})

	profiles, err := svc.List("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}
}

func TestProfileService_List_FilterByGroup(t *testing.T) {
	setupServiceTestDB(t)
	g1 := createTestGroup(t)
	g2 := &database.NodeGroup{UUID: database.GenerateUUID(), Alias: "G2", SortOrder: 20, Enabled: true}
	database.DB.Create(g2)

	svc := NewProfileService()
	database.DB.Create(&database.Profile{UUID: database.GenerateUUID(), Name: "P1", ProxyProtocol: "vless", GroupUUID: g1.UUID, SortOrder: 10})
	database.DB.Create(&database.Profile{UUID: database.GenerateUUID(), Name: "P2", ProxyProtocol: "trojan", GroupUUID: g2.UUID, SortOrder: 10})

	profiles, err := svc.List(g1.UUID, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}
	if profiles[0].Name != "P1" {
		t.Fatalf("expected 'P1', got '%s'", profiles[0].Name)
	}
}

func TestProfileService_Delete(t *testing.T) {
	setupServiceTestDB(t)
	group := createTestGroup(t)

	svc := NewProfileService()
	profile := &database.Profile{
		UUID:          database.GenerateUUID(),
		Name:          "DeleteMe",
		ProxyProtocol: "vless",
		GroupUUID:     group.UUID,
		SortOrder:     10,
	}
	database.DB.Create(profile)

	if err := svc.Delete(profile.UUID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := svc.Get(profile.UUID)
	if err == nil {
		t.Fatal("expected error after deletion")
	}
}

func TestProfileService_Delete_NotFound(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewProfileService()
	err := svc.Delete("nonexistent-uuid")
	if err == nil {
		t.Fatal("expected error for nonexistent profile")
	}
}

func TestProfileService_Update(t *testing.T) {
	setupServiceTestDB(t)
	group := createTestGroup(t)

	svc := NewProfileService()
	profile := &database.Profile{
		UUID:          database.GenerateUUID(),
		Name:          "OldName",
		ProxyProtocol: "vless",
		GroupUUID:     group.UUID,
		SortOrder:     10,
	}
	database.DB.Create(profile)

	updated, err := svc.Update(profile.UUID, map[string]any{
		"name":       "NewName",
		"group_uuid": group.UUID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "NewName" {
		t.Fatalf("expected name 'NewName', got '%s'", updated.Name)
	}
}

func TestProfileService_Select(t *testing.T) {
	setupServiceTestDB(t)
	group := createTestGroup(t)

	svc := NewProfileService()
	p1 := &database.Profile{UUID: database.GenerateUUID(), Name: "P1", ProxyProtocol: "vless", GroupUUID: group.UUID, SortOrder: 10, IsActive: true}
	p2 := &database.Profile{UUID: database.GenerateUUID(), Name: "P2", ProxyProtocol: "vless", GroupUUID: group.UUID, SortOrder: 20}
	database.DB.Create(p1)
	database.DB.Create(p2)

	if err := svc.Select(p2.UUID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// p1 should be deactivated
	var found1 database.Profile
	database.DB.Where("uuid = ?", p1.UUID).First(&found1)
	if found1.IsActive {
		t.Fatal("expected p1 to be deactivated")
	}

	// p2 should be activated
	var found2 database.Profile
	database.DB.Where("uuid = ?", p2.UUID).First(&found2)
	if !found2.IsActive {
		t.Fatal("expected p2 to be activated")
	}
}

func TestProfileService_ImportLinks(t *testing.T) {
	setupServiceTestDB(t)
	group := createTestGroup(t)

	svc := NewProfileService()
	links := "vless://uuid@example.com:443?type=tcp&security=tls#Node1\ntrojan://pass@host.com:443#Node2"

	count, err := svc.ImportLinks(links, group.UUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 imported, got %d", count)
	}

	profiles, _ := svc.List(group.UUID, "")
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles in DB, got %d", len(profiles))
	}
}

func TestProfileService_ImportLinks_NoGroup(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewProfileService()
	_, err := svc.ImportLinks("vless://uuid@host:443#Test", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent group")
	}
}

func TestProfileService_Dedup(t *testing.T) {
	setupServiceTestDB(t)
	group := createTestGroup(t)

	svc := NewProfileService()
	rawLink := "vless://uuid@example.com:443#Node"
	database.DB.Create(&database.Profile{UUID: database.GenerateUUID(), Name: "P1", ProxyProtocol: "vless", GroupUUID: group.UUID, SortOrder: 10, RawLink: rawLink})
	database.DB.Create(&database.Profile{UUID: database.GenerateUUID(), Name: "P2", ProxyProtocol: "vless", GroupUUID: group.UUID, SortOrder: 20, RawLink: rawLink})

	result, err := svc.Dedup("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Removed != 1 {
		t.Fatalf("expected 1 removed, got %d", result.Removed)
	}
	if result.Total != 2 {
		t.Fatalf("expected 2 total, got %d", result.Total)
	}
}

// helper
func isErrValidation(err error, _ *ErrValidation) bool {
	var target *ErrValidation
	if e, ok := err.(*ErrValidation); ok {
		_ = e
		return true
	}
	_ = target
	return false
}
