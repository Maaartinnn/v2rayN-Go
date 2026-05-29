package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"v2rayn-go/database"
	"v2rayn-go/ping"
	"v2rayn-go/service"
)

func setupWebTestDB(t *testing.T) {
	t.Helper()
	database.InitTestDB()
	t.Cleanup(database.CleanupTestDB)
}

func createTestMux(t *testing.T) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()

	profileSvc := service.NewProfileService()
	groupSvc := service.NewGroupService()
	routingSvc := service.NewRoutingRuleService()

	// Create a mock ping service
	pingSvc := &mockPingService{}

	profileHandler := NewProfileHandler(profileSvc, pingSvc)
	groupHandler := NewGroupHandler(groupSvc)
	routingHandler := NewRoutingRuleHandler(routingSvc)

	profileHandler.Register(mux)
	groupHandler.Register(mux)
	routingHandler.Register(mux)

	return mux
}

// mockPingService implements PingServiceInterface for testing
type mockPingService struct{}

func (m *mockPingService) PingSingleProfile(profile *database.Profile) ping.PingResult {
	return ping.PingResult{ProfileUUID: profile.UUID, Latency: 0}
}
func (m *mockPingService) PingAllProfiles(ctx context.Context, concurrency int) []ping.PingResult {
	return nil
}

func TestWebHelper_jsonOK(t *testing.T) {
	w := httptest.NewRecorder()
	jsonOK(w, map[string]string{"status": "ok"})

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var result map[string]string
	json.NewDecoder(w.Body).Decode(&result)
	if result["status"] != "ok" {
		t.Fatalf("expected status 'ok', got '%s'", result["status"])
	}
}

func TestWebHelper_jsonError(t *testing.T) {
	w := httptest.NewRecorder()
	jsonError(w, "not found", http.StatusNotFound)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}

	var result map[string]any
	json.NewDecoder(w.Body).Decode(&result)
	if result["error"] != "not found" {
		t.Fatalf("expected error 'not found', got '%v'", result["error"])
	}
}

// ==================== Profile Handler HTTP Tests ====================

func TestProfileHandler_List(t *testing.T) {
	setupWebTestDB(t)
	mux := createTestMux(t)

	// Create a group and profile
	group := &database.NodeGroup{UUID: database.GenerateUUID(), Alias: "G1", SortOrder: 10, Enabled: true}
	database.DB.Create(group)
	database.DB.Create(&database.Profile{
		UUID: database.GenerateUUID(), Name: "Node1", ProxyProtocol: "vless",
		ProxyAddress: "example.com", ProxyPort: 443,
		GroupUUID: group.UUID, SortOrder: 10,
	})

	req := httptest.NewRequest("GET", "/api/profiles/", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// 列表返回精简的 ProfileListItem（不含 raw_link 等大字段）
	var items []database.ProfileListItem
	json.NewDecoder(w.Body).Decode(&items)
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].UUID == "" {
		t.Fatal("expected UUID to be set")
	}
	if items[0].Name != "Node1" {
		t.Fatalf("expected name 'Node1', got '%s'", items[0].Name)
	}
	if items[0].ProxyProtocol != "vless" {
		t.Fatalf("expected protocol 'vless', got '%s'", items[0].ProxyProtocol)
	}
	// 验证颜色字段已填充
	if items[0].ProtocolColor.Text == "" {
		t.Fatal("expected protocol color to be populated")
	}
	if items[0].LatencyColor == "" {
		t.Fatal("expected latency color to be populated")
	}
}

func TestProfileHandler_List_FilterByGroup(t *testing.T) {
	setupWebTestDB(t)
	mux := createTestMux(t)

	g1 := &database.NodeGroup{UUID: database.GenerateUUID(), Alias: "G1", SortOrder: 10, Enabled: true}
	g2 := &database.NodeGroup{UUID: database.GenerateUUID(), Alias: "G2", SortOrder: 20, Enabled: true}
	database.DB.Create(g1)
	database.DB.Create(g2)
	database.DB.Create(&database.Profile{UUID: database.GenerateUUID(), Name: "P1", ProxyProtocol: "vless", GroupUUID: g1.UUID, SortOrder: 10})
	database.DB.Create(&database.Profile{UUID: database.GenerateUUID(), Name: "P2", ProxyProtocol: "trojan", GroupUUID: g2.UUID, SortOrder: 10})

	req := httptest.NewRequest("GET", "/api/profiles/?group_uuid="+g1.UUID, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// 列表返回精简的 ProfileListItem
	var items []database.ProfileListItem
	json.NewDecoder(w.Body).Decode(&items)
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Name != "P1" {
		t.Fatalf("expected 'P1', got '%s'", items[0].Name)
	}
}

func TestProfileHandler_Get(t *testing.T) {
	setupWebTestDB(t)
	mux := createTestMux(t)

	group := &database.NodeGroup{UUID: database.GenerateUUID(), Alias: "G1", SortOrder: 10, Enabled: true}
	database.DB.Create(group)
	profile := &database.Profile{
		UUID: database.GenerateUUID(), Name: "GetMe", ProxyProtocol: "vless",
		ProxyAddress: "host.com", ProxyPort: 443, GroupUUID: group.UUID, SortOrder: 10,
	}
	database.DB.Create(profile)

	req := httptest.NewRequest("GET", "/api/profiles/"+profile.UUID, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result database.Profile
	json.NewDecoder(w.Body).Decode(&result)
	if result.Name != "GetMe" {
		t.Fatalf("expected name 'GetMe', got '%s'", result.Name)
	}
}

func TestProfileHandler_Get_NotFound(t *testing.T) {
	setupWebTestDB(t)
	mux := createTestMux(t)

	req := httptest.NewRequest("GET", "/api/profiles/nonexistent-uuid", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestProfileHandler_Create(t *testing.T) {
	setupWebTestDB(t)
	mux := createTestMux(t)

	group := &database.NodeGroup{UUID: database.GenerateUUID(), Alias: "G1", SortOrder: 10, Enabled: true}
	database.DB.Create(group)

	body, _ := json.Marshal(map[string]any{
		"name":           "NewNode",
		"proxy_address":  "example.com",
		"proxy_port":     443,
		"proxy_protocol": "vless",
		"group_uuid":     group.UUID,
	})
	req := httptest.NewRequest("POST", "/api/profiles/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result database.Profile
	json.NewDecoder(w.Body).Decode(&result)
	if result.Name != "NewNode" {
		t.Fatalf("expected name 'NewNode', got '%s'", result.Name)
	}
	if result.UUID == "" {
		t.Fatal("expected UUID to be set")
	}
}

func TestProfileHandler_Create_InvalidJSON(t *testing.T) {
	setupWebTestDB(t)
	mux := createTestMux(t)

	req := httptest.NewRequest("POST", "/api/profiles/", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestProfileHandler_Delete(t *testing.T) {
	setupWebTestDB(t)
	mux := createTestMux(t)

	group := &database.NodeGroup{UUID: database.GenerateUUID(), Alias: "G1", SortOrder: 10, Enabled: true}
	database.DB.Create(group)
	profile := &database.Profile{
		UUID: database.GenerateUUID(), Name: "DeleteMe", ProxyProtocol: "vless",
		GroupUUID: group.UUID, SortOrder: 10,
	}
	database.DB.Create(profile)

	req := httptest.NewRequest("DELETE", "/api/profiles/"+profile.UUID, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// ==================== Group Handler HTTP Tests ====================

func TestGroupHandler_List(t *testing.T) {
	setupWebTestDB(t)
	mux := createTestMux(t)

	database.DB.Create(&database.NodeGroup{UUID: database.GenerateUUID(), Alias: "G1", SortOrder: 10, Enabled: true})
	database.DB.Create(&database.NodeGroup{UUID: database.GenerateUUID(), Alias: "G2", SortOrder: 20, Enabled: true})

	req := httptest.NewRequest("GET", "/api/groups/", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var groups []database.NodeGroup
	json.NewDecoder(w.Body).Decode(&groups)
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
}

func TestGroupHandler_Create(t *testing.T) {
	setupWebTestDB(t)
	mux := createTestMux(t)

	body, _ := json.Marshal(map[string]any{
		"alias":   "NewGroup",
		"enabled": true,
	})
	req := httptest.NewRequest("POST", "/api/groups/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result database.NodeGroup
	json.NewDecoder(w.Body).Decode(&result)
	if result.Alias != "NewGroup" {
		t.Fatalf("expected alias 'NewGroup', got '%s'", result.Alias)
	}
}

// ==================== Routing Rule Handler HTTP Tests ====================

func TestRoutingRuleHandler_List(t *testing.T) {
	setupWebTestDB(t)
	mux := createTestMux(t)

	database.DB.Create(&database.RoutingRule{UUID: database.GenerateUUID(), Name: "R1", Type: "direct", Enabled: true, SortOrder: 10})

	req := httptest.NewRequest("GET", "/api/routing-rules/", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var rules []database.RoutingRule
	json.NewDecoder(w.Body).Decode(&rules)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
}

// ==================== Error Mapping Tests ====================

func TestMapServiceError_NotFound(t *testing.T) {
	w := httptest.NewRecorder()
	err := service.NewNotFound("resource missing", nil)
	mapServiceError(w, err)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestMapServiceError_Validation(t *testing.T) {
	w := httptest.NewRecorder()
	err := service.NewValidation("bad input", nil)
	mapServiceError(w, err)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestMapServiceError_Conflict(t *testing.T) {
	w := httptest.NewRecorder()
	err := service.NewConflict("duplicate", nil)
	mapServiceError(w, err)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestMapServiceError_Internal(t *testing.T) {
	w := httptest.NewRecorder()
	err := fmt.Errorf("something went wrong")
	mapServiceError(w, err)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
