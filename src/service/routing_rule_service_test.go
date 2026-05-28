package service

import (
	"testing"
	"v2rayn-go/database"
)

// ==================== RoutingRuleService ====================

func TestRoutingRuleService_Create(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewRoutingRuleService()
	rule := &database.RoutingRule{
		Name:    "Direct CN",
		Type:    "direct",
		Domain:  "cn,geosite:cn",
		Enabled: true,
	}

	if err := svc.Create(rule); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rule.UUID == "" {
		t.Fatal("expected UUID to be set")
	}
	if rule.SortOrder != 10 {
		t.Fatalf("expected sort_order 10, got %d", rule.SortOrder)
	}
}

func TestRoutingRuleService_Get(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewRoutingRuleService()
	rule := &database.RoutingRule{Name: "GetMe", Type: "proxy", Enabled: true}
	svc.Create(rule)

	found, err := svc.Get(rule.UUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.Name != "GetMe" {
		t.Fatalf("expected name 'GetMe', got '%s'", found.Name)
	}
}

func TestRoutingRuleService_Get_NotFound(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewRoutingRuleService()
	_, err := svc.Get("nonexistent-uuid")
	if err == nil {
		t.Fatal("expected error for nonexistent rule")
	}
}

func TestRoutingRuleService_List(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewRoutingRuleService()
	svc.Create(&database.RoutingRule{Name: "R1", Type: "direct", Enabled: true})
	svc.Create(&database.RoutingRule{Name: "R2", Type: "proxy", Enabled: true})

	rules, err := svc.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
}

func TestRoutingRuleService_Update(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewRoutingRuleService()
	rule := &database.RoutingRule{Name: "Before", Type: "direct", Enabled: true}
	svc.Create(rule)

	updated, err := svc.Update(rule.UUID, &database.RoutingRule{
		Name:    "After",
		Type:    "block",
		Enabled: false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "After" {
		t.Fatalf("expected name 'After', got '%s'", updated.Name)
	}
	if updated.Type != "block" {
		t.Fatalf("expected type 'block', got '%s'", updated.Type)
	}
}

func TestRoutingRuleService_Update_NotFound(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewRoutingRuleService()
	_, err := svc.Update("nonexistent", &database.RoutingRule{Name: "test"})
	if err == nil {
		t.Fatal("expected error for nonexistent rule")
	}
}

func TestRoutingRuleService_Delete(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewRoutingRuleService()
	rule := &database.RoutingRule{Name: "DeleteMe", Type: "block", Enabled: true}
	svc.Create(rule)

	if err := svc.Delete(rule.UUID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := svc.Get(rule.UUID)
	if err == nil {
		t.Fatal("expected error after deletion")
	}
}

func TestRoutingRuleService_Delete_NotFound(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewRoutingRuleService()
	err := svc.Delete("nonexistent-uuid")
	if err == nil {
		t.Fatal("expected error for nonexistent rule")
	}
}
