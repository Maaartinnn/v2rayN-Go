package web

import (
	"encoding/json"
	"log"
	"net/http"

	"v2rayn-go/database"
	"v2rayn-go/service"
	"v2rayn-go/subscription"
)

// GroupHandler 分组管理独立处理器
type GroupHandler struct {
	groupSvc *service.GroupService
}

// NewGroupHandler 创建分组管理处理器
func NewGroupHandler(groupSvc *service.GroupService) *GroupHandler {
	return &GroupHandler{groupSvc: groupSvc}
}

// Register 挂载分组管理路由
func (h *GroupHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET    /api/groups/{$}", h.handleList)
	mux.HandleFunc("POST   /api/groups/{$}", h.handleCreate)
	mux.HandleFunc("PUT    /api/groups/reorder", h.handleReorder)

	mux.HandleFunc("GET    /api/groups/{uuid}", h.handleGet)
	mux.HandleFunc("PUT    /api/groups/{uuid}", h.handleUpdate)
	mux.HandleFunc("DELETE /api/groups/{uuid}", h.handleDelete)
	mux.HandleFunc("POST   /api/groups/{uuid}/refresh", h.handleRefresh)
	mux.HandleFunc("POST   /api/groups/{uuid}/refresh-proxy", h.handleRefreshProxy)
}

func (h *GroupHandler) handleList(w http.ResponseWriter, r *http.Request) {
	groups, err := h.groupSvc.List()
	if err != nil {
		mapServiceError(w, err)
		return
	}
	jsonOK(w, groups)
}

func (h *GroupHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var group database.NodeGroup
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if err := h.groupSvc.Create(&group); err != nil {
		mapServiceError(w, err)
		return
	}
	jsonOK(w, group)
}

func (h *GroupHandler) handleReorder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UUID       string `json:"uuid"`
		BeforeUUID string `json:"before_uuid"`
		AfterUUID  string `json:"after_uuid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	newOrder, err := h.groupSvc.Reorder(req.UUID, req.BeforeUUID, req.AfterUUID)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	jsonOK(w, map[string]interface{}{"status": "reordered", "sort_order": newOrder})
}

func (h *GroupHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	group, err := h.groupSvc.Get(uuid)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	jsonOK(w, group)
}

func (h *GroupHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	var updated database.NodeGroup
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	result, err := h.groupSvc.Update(uuid, &updated)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	jsonOK(w, result)
}

func (h *GroupHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if err := h.groupSvc.Delete(uuid); err != nil {
		mapServiceError(w, err)
		return
	}
	jsonOK(w, map[string]string{"status": "deleted"})
}

func (h *GroupHandler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	group, err := h.groupSvc.Get(uuid)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	if !group.IsSubscription {
		jsonError(w, "Group is not a subscription group", http.StatusBadRequest)
		return
	}
	go func() {
		subSvc := subscription.NewService()
		if err := subSvc.UpdateGroupSubscription(group, false); err != nil {
			log.Printf("Failed to refresh group %s: %v", group.Alias, err)
		}
	}()
	jsonOK(w, map[string]string{"status": "refreshing"})
}

func (h *GroupHandler) handleRefreshProxy(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	group, err := h.groupSvc.Get(uuid)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	if !group.IsSubscription {
		jsonError(w, "Group is not a subscription group", http.StatusBadRequest)
		return
	}
	go func() {
		subSvc := subscription.NewService()
		if err := subSvc.UpdateGroupSubscription(group, true); err != nil {
			log.Printf("Failed to refresh group %s via proxy: %v", group.Alias, err)
		}
	}()
	jsonOK(w, map[string]string{"status": "refreshing"})
}
