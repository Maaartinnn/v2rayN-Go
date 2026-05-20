package web

import (
	"log"
	"net/http"
	"sync"

	"v2rayn-go/service"
)

// CoreHandler 核心管理独立处理器
type CoreHandler struct {
	coreSvc     *service.CoreService
	broadcaster StatusBroadcaster
	downloads   *downloadTracker
}

// NewCoreHandler 创建核心管理处理器
func NewCoreHandler(coreSvc *service.CoreService, broadcaster StatusBroadcaster) *CoreHandler {
	return &CoreHandler{
		coreSvc:     coreSvc,
		broadcaster: broadcaster,
		downloads:   newDownloadTracker(),
	}
}

// Register 挂载核心管理路由
func (h *CoreHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/core/start", h.handleCoreStart)
	mux.HandleFunc("POST /api/core/stop", h.handleCoreStop)
	mux.HandleFunc("GET  /api/core/status", h.handleCoreStatus)

	mux.HandleFunc("GET  /api/cores/{$}", h.handleCores)
	mux.HandleFunc("GET  /api/cores/check-updates", h.handleCoresCheckUpdates)
	mux.HandleFunc("GET  /api/cores/detect-versions", h.handleCoresDetectVersions)
	mux.HandleFunc("POST /api/cores/download", h.handleCoreDownload)
	mux.HandleFunc("POST /api/cores/download-url", h.handleCoreDownloadURL)
	mux.HandleFunc("POST /api/cores/upload", h.handleCoreUpload)
}

func (h *CoreHandler) handleCoreStart(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CoreType   string `json:"core_type"`
		ConfigPath string `json:"config_path"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	if err := h.coreSvc.Start(req.CoreType, req.ConfigPath); err != nil {
		mapServiceError(w, err)
		return
	}

	jsonOK(w, map[string]string{"status": "started"})
	go h.broadcaster.BroadcastStatus()
}

func (h *CoreHandler) handleCoreStop(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CoreType string `json:"core_type"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}

	if err := h.coreSvc.Stop(req.CoreType); err != nil {
		mapServiceError(w, err)
		return
	}

	jsonOK(w, map[string]string{"status": "stopped"})
	go h.broadcaster.BroadcastStatus()
}

func (h *CoreHandler) handleCoreStatus(w http.ResponseWriter, r *http.Request) {
	statuses := h.coreSvc.GetAllStatus()
	jsonOK(w, statuses)
}

func (h *CoreHandler) handleCores(w http.ResponseWriter, r *http.Request) {
	cores := h.coreSvc.GetLocalCores()
	jsonOK(w, cores)
}

func (h *CoreHandler) handleCoresCheckUpdates(w http.ResponseWriter, r *http.Request) {
	latestVersions := h.coreSvc.CheckUpdates()
	jsonOK(w, map[string]interface{}{"latest_versions": latestVersions})
}

func (h *CoreHandler) handleCoresDetectVersions(w http.ResponseWriter, r *http.Request) {
	h.coreSvc.DetectVersions(func(versions map[string]string) {
		h.broadcaster.Broadcast(map[string]interface{}{"type": "core_versions", "payload": versions})
	})
	jsonOK(w, map[string]string{"status": "detecting"})
}

func (h *CoreHandler) handleCoreDownload(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CoreName string `json:"core_name"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.CoreName == "" {
		jsonError(w, "Missing core_name", http.StatusBadRequest)
		return
	}
	if h.downloads.Exists(req.CoreName) {
		jsonError(w, "Download already in progress", http.StatusConflict)
		return
	}

	state := h.downloads.Start(req.CoreName)

	go func() {
		defer h.downloads.Delete(req.CoreName)
		err := h.coreSvc.Download(req.CoreName, func(downloaded, total int64) {
			state.Update(downloaded, total)
			h.broadcaster.Broadcast(map[string]interface{}{"type": "download_progress", "payload": state})
		})
		if err != nil {
			state.SetError(err.Error())
			log.Printf("Failed to download core %s: %v", req.CoreName, err)
			h.broadcaster.Broadcast(map[string]interface{}{"type": "download_complete", "payload": map[string]interface{}{"core_name": req.CoreName, "success": false, "error": err.Error()}})
		} else {
			state.SetComplete()
			h.broadcaster.Broadcast(map[string]interface{}{"type": "download_complete", "payload": map[string]interface{}{"core_name": req.CoreName, "success": true}})
		}
	}()

	jsonOK(w, map[string]string{"status": "downloading", "core": req.CoreName})
}

func (h *CoreHandler) handleCoreDownloadURL(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CoreName    string `json:"core_name"`
		DownloadURL string `json:"download_url"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.CoreName == "" || req.DownloadURL == "" {
		jsonError(w, "Missing core_name or download_url", http.StatusBadRequest)
		return
	}
	if h.downloads.Exists(req.CoreName) {
		jsonError(w, "Download already in progress", http.StatusConflict)
		return
	}

	state := h.downloads.Start(req.CoreName)

	go func() {
		defer h.downloads.Delete(req.CoreName)
		err := h.coreSvc.DownloadFromURL(req.CoreName, req.DownloadURL, func(downloaded, total int64) {
			state.Update(downloaded, total)
			h.broadcaster.Broadcast(map[string]interface{}{"type": "download_progress", "payload": state})
		})
		if err != nil {
			state.SetError(err.Error())
			log.Printf("Failed to download core %s from URL: %v", req.CoreName, err)
			h.broadcaster.Broadcast(map[string]interface{}{"type": "download_complete", "payload": map[string]interface{}{"core_name": req.CoreName, "success": false, "error": err.Error()}})
		} else {
			state.SetComplete()
			h.broadcaster.Broadcast(map[string]interface{}{"type": "download_complete", "payload": map[string]interface{}{"core_name": req.CoreName, "success": true}})
		}
	}()

	jsonOK(w, map[string]string{"status": "downloading", "core": req.CoreName, "url": req.DownloadURL})
}

func (h *CoreHandler) handleCoreUpload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(200 << 20); err != nil {
		jsonError(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	coreName := r.FormValue("core_name")
	if coreName == "" {
		jsonError(w, "Missing core_name", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("binary")
	if err != nil {
		jsonError(w, "Missing binary file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	destPath, err := h.coreSvc.Upload(coreName, header.Filename, file)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	jsonOK(w, map[string]string{"status": "uploaded", "core": coreName, "path": destPath})
}

// ========== Download Tracker ==========

// downloadTracker 并发安全的下载状态跟踪器
type downloadTracker struct {
	syncMap sync.Map // map[string]*downloadState
}

func newDownloadTracker() *downloadTracker {
	return &downloadTracker{}
}

func (t *downloadTracker) Exists(name string) bool {
	_, ok := t.syncMap.Load(name)
	return ok
}

func (t *downloadTracker) Start(name string) *downloadState {
	state := &downloadState{CoreName: name, Status: "downloading"}
	t.syncMap.Store(name, state)
	return state
}

func (t *downloadTracker) Delete(name string) {
	t.syncMap.Delete(name)
}

// downloadState 下载状态
type downloadState struct {
	CoreName   string `json:"core_name"`
	Downloaded int64  `json:"downloaded"`
	Total      int64  `json:"total"`
	Percentage int    `json:"percentage"`
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
}

func (s *downloadState) Update(downloaded, total int64) {
	s.Downloaded = downloaded
	s.Total = total
	if total > 0 {
		s.Percentage = int(downloaded * 100 / total)
	}
}

func (s *downloadState) SetError(err string) {
	s.Status = "error"
	s.Error = err
}

func (s *downloadState) SetComplete() {
	s.Status = "complete"
	s.Percentage = 100
}
