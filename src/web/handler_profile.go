package web

import (
	"io"
	"net/http"

	"v2rayn-go/coredef"
	"v2rayn-go/database"
	"v2rayn-go/parser"
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
	mux.HandleFunc("GET    /api/profiles/core-list", h.handleCoreList)
	mux.HandleFunc("POST /api/profiles/import", h.handleImport)
	mux.HandleFunc("POST   /api/profiles/import-image", h.handleImportImage)
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

func (h *ProfileHandler) handleImportImage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(coredef.MultipartMaxMemoryDefault); err != nil {
		jsonError(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	groupUUID := r.FormValue("group_uuid")
	if groupUUID == "" {
		jsonError(w, "group_uuid is required", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		data, err := io.ReadAll(file)
		if err != nil {
			jsonError(w, "Failed to read file", http.StatusInternalServerError)
			return
		}
		links, decodeErr := parser.DecodeQRFromBytes(data)
		if decodeErr != nil {
			jsonError(w, "No QR code found in image: "+decodeErr.Error(), http.StatusBadRequest)
			return
		}
		h.importParsedLinks(w, links, groupUUID)
		return
	}

	imageURL := r.FormValue("url")
	if imageURL == "" {
		jsonError(w, "No image file or URL provided", http.StatusBadRequest)
		return
	}

	resp, err := http.Get(imageURL)
	if err != nil {
		jsonError(w, "Failed to download image: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		jsonError(w, "Failed to read image data", http.StatusInternalServerError)
		return
	}

	links, decodeErr := parser.DecodeQRFromBytes(data)
	if decodeErr != nil {
		jsonError(w, "No QR code found in image: "+decodeErr.Error(), http.StatusBadRequest)
		return
	}

	h.importParsedLinks(w, links, groupUUID)
}

func (h *ProfileHandler) importParsedLinks(w http.ResponseWriter, links []string, groupUUID string) {
	count, err := h.profileSvc.ImportParsedLinks(links, groupUUID)
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

// handleCoreList 处理 GET /api/profiles/core-list?protocol=vmess
//
// 返回指定协议兼容且本地已安装的内核列表（按推荐优先级排序）。
// 用于新增节点时，前端根据当前选中的协议动态获取可用内核列表。
func (h *ProfileHandler) handleCoreList(w http.ResponseWriter, r *http.Request) {
	protocol := r.URL.Query().Get("protocol")
	if protocol == "" {
		jsonError(w, "protocol is required", http.StatusBadRequest)
		return
	}

	var coreList []string
	if h.coreSvc != nil {
		coreList = h.coreSvc.GetCompatibleInstalledCores(protocol)
	}
	jsonOK(w, map[string]any{"core_list": coreList})
}

// handleGet 处理 GET /api/profiles/{uuid}，返回完整节点数据 + core_list。
//
// core_list 是该协议兼容且本地已安装的内核列表（按推荐优先级排序），
// 由 CoreService.GetCompatibleInstalledCores() 计算。
// 前端 NodeEditForm 直接使用 core_list 渲染内核选择下拉框，
// 不再需要单独调用 GET /api/cores 获取所有内核再做前端过滤。
func (h *ProfileHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	profile, err := h.profileSvc.Get(uuid)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	// coreSvc 可能为 nil（测试环境），此时 core_list 返回 nil
	var coreList []string
	if h.coreSvc != nil {
		coreList = h.coreSvc.GetCompatibleInstalledCores(profile.ProxyProtocol)
	}
	jsonOK(w, map[string]any{
		"profile":   profile,
		"core_list": coreList,
	})
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
