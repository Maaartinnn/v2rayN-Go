package database

import (
	"testing"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	InitTestDB()
	t.Cleanup(CleanupTestDB)
}

// ==================== NodeGroup CRUD ====================

func TestNodeGroup_Create(t *testing.T) {
	setupTestDB(t)

	group := &NodeGroup{
		UUID:      GenerateUUID(),
		Alias:     "Test Group",
		SortOrder: 10,
		Enabled:   true,
	}
	if err := DB.Create(group).Error; err != nil {
		t.Fatalf("failed to create group: %v", err)
	}
	if group.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}
}

func TestNodeGroup_Get(t *testing.T) {
	setupTestDB(t)

	uuid := GenerateUUID()
	group := &NodeGroup{
		UUID:      uuid,
		Alias:     "My Group",
		SortOrder: 10,
		Enabled:   true,
	}
	DB.Create(group)

	var found NodeGroup
	if err := DB.Where("uuid = ?", uuid).First(&found).Error; err != nil {
		t.Fatalf("failed to get group: %v", err)
	}
	if found.Alias != "My Group" {
		t.Fatalf("expected alias 'My Group', got '%s'", found.Alias)
	}
}

func TestNodeGroup_Update(t *testing.T) {
	setupTestDB(t)

	group := &NodeGroup{
		UUID:      GenerateUUID(),
		Alias:     "Before",
		SortOrder: 10,
		Enabled:   true,
	}
	DB.Create(group)

	DB.Model(group).Update("alias", "After")

	var found NodeGroup
	DB.Where("uuid = ?", group.UUID).First(&found)
	if found.Alias != "After" {
		t.Fatalf("expected alias 'After', got '%s'", found.Alias)
	}
}

func TestNodeGroup_SoftDelete(t *testing.T) {
	setupTestDB(t)

	group := &NodeGroup{
		UUID:      GenerateUUID(),
		Alias:     "ToDelete",
		SortOrder: 10,
		Enabled:   true,
	}
	DB.Create(group)

	DB.Delete(group)

	var count int64
	DB.Model(&NodeGroup{}).Where("uuid = ?", group.UUID).Count(&count)
	if count != 0 {
		t.Fatal("expected 0 records after soft delete")
	}

	// With Unscoped, should still see deleted
	DB.Unscoped().Model(&NodeGroup{}).Where("uuid = ?", group.UUID).Count(&count)
	if count != 1 {
		t.Fatalf("expected 1 record with Unscoped, got %d", count)
	}
}

// ==================== Profile CRUD ====================

func TestProfile_Create(t *testing.T) {
	setupTestDB(t)

	group := &NodeGroup{UUID: GenerateUUID(), Alias: "G1", SortOrder: 10, Enabled: true}
	DB.Create(group)

	profile := &Profile{
		UUID:          GenerateUUID(),
		Name:          "TestNode",
		ProxyAddress:  "example.com",
		ProxyPort:     443,
		ProxyProtocol: "vless",
		GroupUUID:     group.UUID,
		SortOrder:     10,
	}
	if err := DB.Create(profile).Error; err != nil {
		t.Fatalf("failed to create profile: %v", err)
	}
	if profile.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}
}

func TestProfile_ListByGroup(t *testing.T) {
	setupTestDB(t)

	group := &NodeGroup{UUID: GenerateUUID(), Alias: "G1", SortOrder: 10, Enabled: true}
	DB.Create(group)

	p1 := &Profile{UUID: GenerateUUID(), Name: "Node1", ProxyProtocol: "vless", GroupUUID: group.UUID, SortOrder: 10}
	p2 := &Profile{UUID: GenerateUUID(), Name: "Node2", ProxyProtocol: "trojan", GroupUUID: group.UUID, SortOrder: 20}
	DB.Create(p1)
	DB.Create(p2)

	var profiles []Profile
	DB.Where("group_uuid = ?", group.UUID).Order("sort_order ASC").Find(&profiles)
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}
	if profiles[0].Name != "Node1" {
		t.Fatalf("expected first profile 'Node1', got '%s'", profiles[0].Name)
	}
}

func TestProfile_SoftDelete(t *testing.T) {
	setupTestDB(t)

	group := &NodeGroup{UUID: GenerateUUID(), Alias: "G1", SortOrder: 10, Enabled: true}
	DB.Create(group)

	profile := &Profile{UUID: GenerateUUID(), Name: "DeleteMe", ProxyProtocol: "vless", GroupUUID: group.UUID, SortOrder: 10}
	DB.Create(profile)

	DB.Delete(profile)

	var count int64
	DB.Model(&Profile{}).Where("uuid = ?", profile.UUID).Count(&count)
	if count != 0 {
		t.Fatal("expected 0 after soft delete")
	}
}

// ==================== RoutingRule CRUD ====================

func TestRoutingRule_Create(t *testing.T) {
	setupTestDB(t)

	rule := &RoutingRule{
		UUID:      GenerateUUID(),
		Name:      "Direct CN",
		Type:      "direct",
		Domain:    "cn,geosite:cn",
		Enabled:   true,
		SortOrder: 10,
	}
	if err := DB.Create(rule).Error; err != nil {
		t.Fatalf("failed to create routing rule: %v", err)
	}
	if rule.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
}

// ==================== StrategyGroup CRUD ====================

func TestStrategyGroup_Create(t *testing.T) {
	setupTestDB(t)

	sg := &StrategyGroup{
		UUID:         GenerateUUID(),
		Name:         "AutoSelect",
		Type:         "urltest",
		TestURL:      "https://www.gstatic.com/generate_204",
		TestInterval: 300,
		Enabled:      true,
		SortOrder:    10,
	}
	if err := DB.Create(sg).Error; err != nil {
		t.Fatalf("failed to create strategy group: %v", err)
	}
	if sg.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
}

func TestStrategyGroup_UniqueName(t *testing.T) {
	setupTestDB(t)

	sg1 := &StrategyGroup{UUID: GenerateUUID(), Name: "Unique", Type: "selector", SortOrder: 10, Enabled: true}
	DB.Create(sg1)

	sg2 := &StrategyGroup{UUID: GenerateUUID(), Name: "Unique", Type: "selector", SortOrder: 20, Enabled: true}
	err := DB.Create(sg2).Error
	if err == nil {
		t.Fatal("expected error for duplicate strategy group name")
	}
}

// ==================== AppSetting CRUD ====================

func TestAppSetting_CreateAndGet(t *testing.T) {
	setupTestDB(t)

	setting := &AppSetting{Key: "theme", Value: "dark"}
	DB.Create(setting)

	var found AppSetting
	if err := DB.Where("key = ?", "theme").First(&found).Error; err != nil {
		t.Fatalf("failed to get setting: %v", err)
	}
	if found.Value != "dark" {
		t.Fatalf("expected value 'dark', got '%s'", found.Value)
	}
}

func TestAppSetting_UniqueKey(t *testing.T) {
	setupTestDB(t)

	s1 := &AppSetting{Key: "lang", Value: "en"}
	DB.Create(s1)

	s2 := &AppSetting{Key: "lang", Value: "zh"}
	err := DB.Create(s2).Error
	if err == nil {
		t.Fatal("expected error for duplicate key")
	}
}

func TestAppSetting_Update(t *testing.T) {
	setupTestDB(t)

	setting := &AppSetting{Key: "port", Value: "1080"}
	DB.Create(setting)

	DB.Model(setting).Update("value", "8080")

	var found AppSetting
	DB.Where("key = ?", "port").First(&found)
	if found.Value != "8080" {
		t.Fatalf("expected '8080', got '%s'", found.Value)
	}
}

// ==================== Sort Utils with DB ====================

func TestSortNewScoped(t *testing.T) {
	setupTestDB(t)

	group := &NodeGroup{UUID: GenerateUUID(), Alias: "G1", SortOrder: 10, Enabled: true}
	DB.Create(group)

	p1 := &Profile{UUID: GenerateUUID(), Name: "P1", GroupUUID: group.UUID, SortOrder: 10}
	p2 := &Profile{UUID: GenerateUUID(), Name: "P2", GroupUUID: group.UUID, SortOrder: 20}
	DB.Create(p1)
	DB.Create(p2)

	newOrder := SortNewScoped(&Profile{}, "group_uuid = ?", group.UUID)
	if newOrder != 30 {
		t.Fatalf("expected new sort_order 30, got %d", newOrder)
	}
}

func TestSortNew_EmptyTable(t *testing.T) {
	setupTestDB(t)

	newOrder := SortNew(&NodeGroup{})
	if newOrder != SortStep {
		t.Fatalf("expected %d for empty table, got %d", SortStep, newOrder)
	}
}

func TestRebalanceScoped(t *testing.T) {
	setupTestDB(t)

	group := &NodeGroup{UUID: GenerateUUID(), Alias: "G1", SortOrder: 10, Enabled: true}
	DB.Create(group)

	// Create profiles with irregular sort orders
	DB.Create(&Profile{UUID: GenerateUUID(), Name: "P1", GroupUUID: group.UUID, SortOrder: 5})
	DB.Create(&Profile{UUID: GenerateUUID(), Name: "P2", GroupUUID: group.UUID, SortOrder: 99})
	DB.Create(&Profile{UUID: GenerateUUID(), Name: "P3", GroupUUID: group.UUID, SortOrder: 3})

	changed := RebalanceScoped(&Profile{}, "group_uuid = ?", group.UUID)
	if !changed {
		t.Fatal("expected rebalance to detect changes")
	}

	var profiles []Profile
	DB.Where("group_uuid = ?", group.UUID).Order("sort_order ASC").Find(&profiles)
	expected := []int{10, 20, 30}
	for i, p := range profiles {
		if p.SortOrder != expected[i] {
			t.Fatalf("profile[%d] sort_order: expected %d, got %d", i, expected[i], p.SortOrder)
		}
	}
}

func TestRebalance_NoChange(t *testing.T) {
	setupTestDB(t)

	group := &NodeGroup{UUID: GenerateUUID(), Alias: "G1", SortOrder: 10, Enabled: true}
	DB.Create(group)

	DB.Create(&Profile{UUID: GenerateUUID(), Name: "P1", GroupUUID: group.UUID, SortOrder: 10})
	DB.Create(&Profile{UUID: GenerateUUID(), Name: "P2", GroupUUID: group.UUID, SortOrder: 20})

	changed := RebalanceScoped(&Profile{}, "group_uuid = ?", group.UUID)
	if changed {
		t.Fatal("expected no rebalance needed")
	}
}

// ==================== GenerateUUID ====================

func TestGenerateUUID_Uniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		u := GenerateUUID()
		if u == "" {
			t.Fatal("expected non-empty UUID")
		}
		if seen[u] {
			t.Fatalf("duplicate UUID: %s", u)
		}
		seen[u] = true
	}
}
