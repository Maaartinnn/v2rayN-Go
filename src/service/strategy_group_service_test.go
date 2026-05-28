package service

import (
	"testing"
	"v2rayn-go/database"
)

// ==================== StrategyGroupService ====================

func TestStrategyGroupService_Create(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewStrategyGroupService()
	group := &database.StrategyGroup{
		Name:         "AutoSelect",
		Type:         "urltest",
		TestURL:      "https://www.gstatic.com/generate_204",
		TestInterval: 300,
		Enabled:      true,
	}

	if err := svc.Create(group); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if group.UUID == "" {
		t.Fatal("expected UUID to be set")
	}
	if group.SortOrder != 10 {
		t.Fatalf("expected sort_order 10, got %d", group.SortOrder)
	}
}

func TestStrategyGroupService_Get(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewStrategyGroupService()
	group := &database.StrategyGroup{Name: "GetMe", Type: "selector", Enabled: true}
	svc.Create(group)

	found, err := svc.Get(group.UUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.Name != "GetMe" {
		t.Fatalf("expected name 'GetMe', got '%s'", found.Name)
	}
}

func TestStrategyGroupService_Get_NotFound(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewStrategyGroupService()
	_, err := svc.Get("nonexistent-uuid")
	if err == nil {
		t.Fatal("expected error for nonexistent strategy group")
	}
}

func TestStrategyGroupService_List(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewStrategyGroupService()
	svc.Create(&database.StrategyGroup{Name: "SG1", Type: "selector", Enabled: true})
	svc.Create(&database.StrategyGroup{Name: "SG2", Type: "urltest", Enabled: true})

	groups, err := svc.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 strategy groups, got %d", len(groups))
	}
}

func TestStrategyGroupService_Update(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewStrategyGroupService()
	group := &database.StrategyGroup{Name: "Before", Type: "selector", Enabled: true}
	svc.Create(group)

	updated, err := svc.Update(group.UUID, &database.StrategyGroup{
		Name:    "After",
		Type:    "urltest",
		Enabled: false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "After" {
		t.Fatalf("expected name 'After', got '%s'", updated.Name)
	}
	if updated.Type != "urltest" {
		t.Fatalf("expected type 'urltest', got '%s'", updated.Type)
	}
}

func TestStrategyGroupService_Update_NotFound(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewStrategyGroupService()
	_, err := svc.Update("nonexistent", &database.StrategyGroup{Name: "test"})
	if err == nil {
		t.Fatal("expected error for nonexistent strategy group")
	}
}

func TestStrategyGroupService_Delete(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewStrategyGroupService()
	group := &database.StrategyGroup{Name: "DeleteMe", Type: "selector", Enabled: true}
	svc.Create(group)

	if err := svc.Delete(group.UUID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := svc.Get(group.UUID)
	if err == nil {
		t.Fatal("expected error after deletion")
	}
}

func TestStrategyGroupService_Delete_NotFound(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewStrategyGroupService()
	err := svc.Delete("nonexistent-uuid")
	if err == nil {
		t.Fatal("expected error for nonexistent strategy group")
	}
}
