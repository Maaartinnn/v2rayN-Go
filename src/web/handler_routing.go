package web

import (
	"net/http"

	"v2rayn-go/database"
	"v2rayn-go/service"
)

// RoutingRuleHandler 路由规则管理独立处理器
type RoutingRuleHandler struct {
	routingSvc *service.RoutingRuleService
}

// NewRoutingRuleHandler 创建路由规则管理处理器
func NewRoutingRuleHandler(routingSvc *service.RoutingRuleService) *RoutingRuleHandler {
	return &RoutingRuleHandler{routingSvc: routingSvc}
}

// Register 挂载路由规则管理路由
func (h *RoutingRuleHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET    /api/routing-rules/{$}", h.handleList)
	mux.HandleFunc("POST   /api/routing-rules/{$}", h.handleCreate)
	mux.HandleFunc("PUT    /api/routing-rules/reorder", h.handleReorder)

	mux.HandleFunc("PUT    /api/routing-rules/{uuid}", h.handleUpdate)
	mux.HandleFunc("DELETE /api/routing-rules/{uuid}", h.handleDelete)
}

func (h *RoutingRuleHandler) handleList(w http.ResponseWriter, r *http.Request) {
	rules, err := h.routingSvc.List()
	if err != nil {
		mapServiceError(w, err)
		return
	}
	jsonOK(w, rules)
}

func (h *RoutingRuleHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var rule database.RoutingRule
	if !decodeJSON(w, r, &rule) {
		return
	}
	if err := h.routingSvc.Create(&rule); err != nil {
		mapServiceError(w, err)
		return
	}
	jsonOK(w, rule)
}

func (h *RoutingRuleHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	var updated database.RoutingRule
	if !decodeJSON(w, r, &updated) {
		return
	}
	result, err := h.routingSvc.Update(uuid, &updated)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	jsonOK(w, result)
}

func (h *RoutingRuleHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if err := h.routingSvc.Delete(uuid); err != nil {
		mapServiceError(w, err)
		return
	}
	jsonOK(w, map[string]string{"status": "deleted"})
}

func (h *RoutingRuleHandler) handleReorder(w http.ResponseWriter, r *http.Request) {
	var req ReorderRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	newOrder, err := h.routingSvc.Reorder(req.UUID, req.BeforeUUID, req.AfterUUID)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	jsonOK(w, map[string]interface{}{"status": "reordered", "sort_order": newOrder})
}
