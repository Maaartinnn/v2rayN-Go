package web

import (
	"net/http"

	"v2rayn-go/coredef"
	"v2rayn-go/database"
	"v2rayn-go/service"
)

// ProfileHandler 节点管理独立处理器
type ProfileHandler struct {
	profileSvc *service.ProfileService
	coreSvc    *service.CoreService
	pingSvc    service.PingServiceInterface
}

// NewProfileHandler 创建节点管理处理器
func NewProfileHandler(profileSvc *service.ProfileService, coreSvc *service.CoreService, pingSvc service.PingServiceInterface) *ProfileHandler {
	return &ProfileHandler{
		profileSvc: profileSvc,
		coreSvc:    coreSvc,
		pingSvc:    pingSvc,
	}
}

// Register 挂载节点管理路由
func (h *ProfileHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET    /api/profiles/{$}", h.handleList)
	mux.HandleFunc("POST   /api/profiles/{$}", h.handleCreate)
	mux.HandleFunc("GET    /api/profiles/core-matrix", h.handleCoreMatrix)
	mux.HandleFunc("POST /api/profiles/import", h.handleImport)
	mux.HandleFunc("POST   /api/profiles/dedup", h.handleDedup)
	mux.HandleFunc("POST   /api/profiles/ping-all", h.handlePingAll)

	mux.HandleFunc("GET    /api/profiles/{uuid}", h.handleGet)
	mux.HandleFunc("PUT    /api/profiles/{uuid}", h.handleUpdate)
	mux.HandleFunc("DELETE /api/profiles/{uuid}", h.handleDelete)
	mux.HandleFunc("POST   /api/profiles/{uuid}/select", h.handleSelect)
	mux.HandleFunc("POST   /api/profiles/{uuid}/ping", h.handlePing)
}

// handleList 处理 GET /api/profiles，支持 group_uuid 和 q 查询参数进行服务端筛选。
// 返回精简的 ProfileListItem 列表（仅展示字段 + 后端计算颜色），减少传输数据量。
// 编辑节点时通过 GET /api/profiles/{uuid} 按需获取完整数据。
func (h *ProfileHandler) handleList(w http.ResponseWriter, r *http.Request) {
	groupUUID := r.URL.Query().Get("group_uuid")
	q := r.URL.Query().Get("q")
	items, err := h.profileSvc.ListSummary(groupUUID, q)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	jsonOK(w, items)
}

func (h *ProfileHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var profile database.Profile
	if !decodeJSON(w, r, &profile) {
		return
	}
	if err := h.profileSvc.Create(&profile); err != nil {
		mapServiceError(w, err)
		return
	}
	jsonOK(w, profile)
}

func (h *ProfileHandler) handleImport(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Links     string `json:"links"`
		GroupUUID string `json:"group_uuid"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	count, err := h.profileSvc.ImportLinks(req.Links, req.GroupUUID)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	jsonOK(w, map[string]int{"imported": count})
}

func (h *ProfileHandler) handleDedup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		GroupUUID string `json:"group_uuid"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	result, err := h.profileSvc.Dedup(req.GroupUUID)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	jsonOK(w, map[string]any{
		"removed": result.Removed,
		"total":   result.Total,
	})
}

func (h *ProfileHandler) handlePingAll(w http.ResponseWriter, r *http.Request) {
	go h.pingSvc.PingAllProfiles(r.Context(), coredef.PingAllConcurrency)
	jsonOK(w, map[string]string{"status": "pinging"})
}

// handleCoreMatrix 处理 GET /api/profiles/core-matrix
//
// 返回当前环境所有协议对应的可用内核矩阵（能力矩阵）。
// 格式：{"vmess": ["xray", "sing-box", "mihomo"], "anytls": ["sing-box"]}
//
// 用于新增节点时，前端一次性获取所有协议的兼容性数据，
// 切换协议时直接查字典，零延迟，无需再次请求后端。
func (h *ProfileHandler) handleCoreMatrix(w http.ResponseWriter, r *http.Request) {
	var matrix map[string][]string
	if h.coreSvc != nil {
		matrix = h.coreSvc.GetInstalledCoreMatrix()
	}
	jsonOK(w, map[string]any{"core_matrix": matrix})
}

// handleGet 处理 GET /api/profiles/{uuid}，返回完整节点数据。
//
// core_matrix 不在此端点返回，前端通过 GET /api/profiles/core-matrix 单独获取。
// 这样保持 API 职责单一，前端逻辑也更清晰。
func (h *ProfileHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	profile, err := h.profileSvc.Get(uuid)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	jsonOK(w, profile)
}

func (h *ProfileHandler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	var req map[string]any
	if !decodeJSON(w, r, &req) {
		return
	}
	profile, err := h.profileSvc.Update(uuid, req)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	jsonOK(w, profile)
}

func (h *ProfileHandler) handleDelete(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if err := h.profileSvc.Delete(uuid); err != nil {
		mapServiceError(w, err)
		return
	}
	jsonOK(w, map[string]string{"status": "deleted"})
}

func (h *ProfileHandler) handleSelect(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if err := h.profileSvc.Select(uuid); err != nil {
		mapServiceError(w, err)
		return
	}
	jsonOK(w, map[string]string{"status": "selected"})
}

func (h *ProfileHandler) handlePing(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	profile, err := h.profileSvc.Get(uuid)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	go h.pingSvc.PingSingleProfile(profile)
	jsonOK(w, map[string]string{"status": "pinging"})
}
