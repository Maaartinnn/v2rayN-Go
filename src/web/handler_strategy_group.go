package web

import (
	"net/http"

	"v2rayn-go/database"
	"v2rayn-go/service"
)

// StrategyGroupHandler 策略组管理独立处理器
type StrategyGroupHandler struct {
	strategySvc *service.StrategyGroupService
}

// NewStrategyGroupHandler 创建策略组管理处理器
func NewStrategyGroupHandler(strategySvc *service.StrategyGroupService) *StrategyGroupHandler {
	return &StrategyGroupHandler{strategySvc: strategySvc}
}

// Register 挂载策略组管理路由
func (h *StrategyGroupHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET    /api/strategy-groups/{$}", h.handleList)
	mux.HandleFunc("POST   /api/strategy-groups/{$}", h.handleCreate)

	mux.HandleFunc("GET    /api/strategy-groups/{uuid}", h.handleGet)
	mux.HandleFunc("PUT    /api/strategy-groups/{uuid}", h.handleUpdate)
	mux.HandleFunc("DELETE /api/strategy-groups/{uuid}", h.handleDelete)
}

func (h *StrategyGroupHandler) handleList(w http.ResponseWriter, r *http.Request) {
	groups, err := h.strategySvc.List()
	if err != nil {
		mapServiceError(w, err)
		return
	}
	jsonOK(w, groups)
}

func (h *StrategyGroupHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var group database.StrategyGroup
	if !decodeJSON(w, r, &group) {
		return
	}
	if err := h.strategySvc.Create(&group); err != nil {
		mapServiceError(w, err)
		return
	}
	jsonOK(w, group)
}

func (h *StrategyGroupHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	group, err := h.strategySvc.Get(uuid)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	jsonOK(w, group)
}

func (h *StrategyGroupHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	var updated database.StrategyGroup
	if !decodeJSON(w, r, &updated) {
		return
	}
	result, err := h.strategySvc.Update(uuid, &updated)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	jsonOK(w, result)
}

func (h *StrategyGroupHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if err := h.strategySvc.Delete(uuid); err != nil {
		mapServiceError(w, err)
		return
	}
	jsonOK(w, map[string]string{"status": "deleted"})
}
