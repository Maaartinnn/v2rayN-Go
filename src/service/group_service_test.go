package service

import (
	"testing"
	"v2rayn-go/database"
)

// ==================== GroupService ====================

func TestGroupService_Create(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewGroupService()
	group := &database.NodeGroup{
		Alias:   "My Group",
		Enabled: true,
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

func TestGroupService_Create_WithUUID(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewGroupService()
	uuid := database.GenerateUUID()
	group := &database.NodeGroup{
		UUID:    uuid,
		Alias:   "WithUUID",
		Enabled: true,
	}

	if err := svc.Create(group); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if group.UUID != uuid {
		t.Fatalf("expected UUID '%s', got '%s'", uuid, group.UUID)
	}
}

func TestGroupService_Get(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewGroupService()
	group := &database.NodeGroup{Alias: "GetMe", Enabled: true}
	svc.Create(group)

	found, err := svc.Get(group.UUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.Alias != "GetMe" {
		t.Fatalf("expected alias 'GetMe', got '%s'", found.Alias)
	}
}

func TestGroupService_Get_NotFound(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewGroupService()
	_, err := svc.Get("nonexistent-uuid")
	if err == nil {
		t.Fatal("expected error for nonexistent group")
	}
}

func TestGroupService_List(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewGroupService()
	svc.Create(&database.NodeGroup{Alias: "G1", Enabled: true})
	svc.Create(&database.NodeGroup{Alias: "G2", Enabled: true})

	groups, err := svc.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
}

func TestGroupService_List_NodeCount(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewGroupService()
	group := &database.NodeGroup{Alias: "WithNodes", Enabled: true}
	svc.Create(group)

	database.DB.Create(&database.Profile{UUID: database.GenerateUUID(), Name: "P1", ProxyProtocol: "vless", GroupUUID: group.UUID, SortOrder: 10})
	database.DB.Create(&database.Profile{UUID: database.GenerateUUID(), Name: "P2", ProxyProtocol: "vless", GroupUUID: group.UUID, SortOrder: 20})

	groups, _ := svc.List()
	for _, g := range groups {
		if g.UUID == group.UUID {
			if g.NodeCount != 2 {
				t.Fatalf("expected node_count 2, got %d", g.NodeCount)
			}
			return
		}
	}
	t.Fatal("expected to find group in list")
}

func TestGroupService_Update(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewGroupService()
	group := &database.NodeGroup{Alias: "Before", Enabled: true}
	svc.Create(group)

	updated, err := svc.Update(group.UUID, &database.NodeGroup{
		Alias:   "After",
		Enabled: false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Alias != "After" {
		t.Fatalf("expected alias 'After', got '%s'", updated.Alias)
	}
	if updated.Enabled != false {
		t.Fatal("expected enabled=false")
	}
}

func TestGroupService_Update_NotFound(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewGroupService()
	_, err := svc.Update("nonexistent", &database.NodeGroup{Alias: "test"})
	if err == nil {
		t.Fatal("expected error for nonexistent group")
	}
}

func TestGroupService_Delete(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewGroupService()
	// Create 2 groups (cannot delete the last one)
	svc.Create(&database.NodeGroup{Alias: "Keep", Enabled: true})
	group2 := &database.NodeGroup{Alias: "Delete", Enabled: true}
	svc.Create(group2)

	if err := svc.Delete(group2.UUID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	groups, _ := svc.List()
	if len(groups) != 1 {
		t.Fatalf("expected 1 group after delete, got %d", len(groups))
	}
}

func TestGroupService_Delete_LastGroup(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewGroupService()
	group := &database.NodeGroup{Alias: "Only", Enabled: true}
	svc.Create(group)

	err := svc.Delete(group.UUID)
	if err == nil {
		t.Fatal("expected error when deleting last group")
	}
	var conflictErr *ErrConflict
	if !isErrConflict(err, conflictErr) {
		t.Fatalf("expected ErrConflict, got %T: %v", err, err)
	}
}

func TestGroupService_Delete_CascadeProfiles(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewGroupService()
	svc.Create(&database.NodeGroup{Alias: "Keep", Enabled: true})
	group2 := &database.NodeGroup{Alias: "Delete", Enabled: true}
	svc.Create(group2)

	database.DB.Create(&database.Profile{UUID: database.GenerateUUID(), Name: "P1", ProxyProtocol: "vless", GroupUUID: group2.UUID, SortOrder: 10})

	if err := svc.Delete(group2.UUID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var profileCount int64
	database.DB.Model(&database.Profile{}).Where("group_uuid = ?", group2.UUID).Count(&profileCount)
	if profileCount != 0 {
		t.Fatalf("expected 0 profiles after cascade delete, got %d", profileCount)
	}
}

func TestGroupService_Delete_NotFound(t *testing.T) {
	setupServiceTestDB(t)

	svc := NewGroupService()
	err := svc.Delete("nonexistent-uuid")
	if err == nil {
		t.Fatal("expected error for nonexistent group")
	}
}

// helper
func isErrConflict(err error, _ *ErrConflict) bool {
	_, ok := err.(*ErrConflict)
	return ok
}
