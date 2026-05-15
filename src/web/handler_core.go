package web

import (
	"encoding/json"
	"log"
	"net/http"
)

// RegisterCoreRoutes 注册核心管理相关路由
func (s *Server) RegisterCoreRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/core/start", s.handleCoreStart)
	mux.HandleFunc("POST /api/core/stop", s.handleCoreStop)
	mux.HandleFunc("GET  /api/core/status", s.handleCoreStatus)

	mux.HandleFunc("GET  /api/cores/{$}", s.handleCores)
	mux.HandleFunc("GET  /api/cores/check-updates", s.handleCoresCheckUpdates)
	mux.HandleFunc("GET  /api/cores/detect-versions", s.handleCoresDetectVersions)
	mux.HandleFunc("POST /api/cores/download{$}", s.handleCoreDownload)
	mux.HandleFunc("POST /api/cores/download-url", s.handleCoreDownloadURL)
	mux.HandleFunc("POST /api/cores/upload", s.handleCoreUpload)
}

func (s *Server) handleCoreStart(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CoreType   string `json:"core_type"`
		ConfigPath string `json:"config_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := s.coreSvc.Start(req.CoreType, req.ConfigPath); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]string{"status": "started"})
	go s.broadcastStatus()
}

func (s *Server) handleCoreStop(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CoreType string `json:"core_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := s.coreSvc.Stop(req.CoreType); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]string{"status": "stopped"})
	go s.broadcastStatus()
}

func (s *Server) handleCoreStatus(w http.ResponseWriter, r *http.Request) {
	statuses := s.coreSvc.GetAllStatus()
	jsonOK(w, statuses)
}

func (s *Server) handleCores(w http.ResponseWriter, r *http.Request) {
	cores := s.coreSvc.GetLocalCores()
	jsonOK(w, cores)
}

func (s *Server) handleCoresCheckUpdates(w http.ResponseWriter, r *http.Request) {
	latestVersions := s.coreSvc.CheckUpdates()
	jsonOK(w, map[string]interface{}{"latest_versions": latestVersions})
}

func (s *Server) handleCoresDetectVersions(w http.ResponseWriter, r *http.Request) {
	s.coreSvc.DetectVersions(func(versions map[string]string) {
		s.broadcastToAll(map[string]interface{}{"type": "core_versions", "payload": versions})
	})
	jsonOK(w, map[string]string{"status": "detecting"})
}

func (s *Server) handleCoreDownload(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CoreName string `json:"core_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if req.CoreName == "" {
		jsonError(w, "Missing core_name", http.StatusBadRequest)
		return
	}
	if _, exists := s.activeDownloads.Load(req.CoreName); exists {
		jsonError(w, "Download already in progress", http.StatusConflict)
		return
	}

	state := &downloadState{CoreName: req.CoreName, Status: "downloading"}
	s.activeDownloads.Store(req.CoreName, state)

	go func() {
		defer s.activeDownloads.Delete(req.CoreName)
		err := s.coreSvc.Download(req.CoreName, func(downloaded, total int64) {
			state.Downloaded = downloaded
			state.Total = total
			if total > 0 {
				state.Percentage = int(downloaded * 100 / total)
			}
			s.broadcastToAll(map[string]interface{}{"type": "download_progress", "payload": state})
		})
		if err != nil {
			state.Status = "error"
			state.Error = err.Error()
			log.Printf("Failed to download core %s: %v", req.CoreName, err)
			s.broadcastToAll(map[string]interface{}{"type": "download_complete", "payload": map[string]interface{}{"core_name": req.CoreName, "success": false, "error": err.Error()}})
		} else {
			state.Status = "complete"
			state.Percentage = 100
			s.broadcastToAll(map[string]interface{}{"type": "download_complete", "payload": map[string]interface{}{"core_name": req.CoreName, "success": true}})
		}
	}()

	jsonOK(w, map[string]string{"status": "downloading", "core": req.CoreName})
}

func (s *Server) handleCoreDownloadURL(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CoreName    string `json:"core_name"`
		DownloadURL string `json:"download_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if req.CoreName == "" || req.DownloadURL == "" {
		jsonError(w, "Missing core_name or download_url", http.StatusBadRequest)
		return
	}
	if _, exists := s.activeDownloads.Load(req.CoreName); exists {
		jsonError(w, "Download already in progress", http.StatusConflict)
		return
	}

	state := &downloadState{CoreName: req.CoreName, Status: "downloading"}
	s.activeDownloads.Store(req.CoreName, state)

	go func() {
		defer s.activeDownloads.Delete(req.CoreName)
		err := s.coreSvc.DownloadFromURL(req.CoreName, req.DownloadURL, func(downloaded, total int64) {
			state.Downloaded = downloaded
			state.Total = total
			if total > 0 {
				state.Percentage = int(downloaded * 100 / total)
			}
			s.broadcastToAll(map[string]interface{}{"type": "download_progress", "payload": state})
		})
		if err != nil {
			state.Status = "error"
			state.Error = err.Error()
			log.Printf("Failed to download core %s from URL: %v", req.CoreName, err)
			s.broadcastToAll(map[string]interface{}{"type": "download_complete", "payload": map[string]interface{}{"core_name": req.CoreName, "success": false, "error": err.Error()}})
		} else {
			state.Status = "complete"
			state.Percentage = 100
			s.broadcastToAll(map[string]interface{}{"type": "download_complete", "payload": map[string]interface{}{"core_name": req.CoreName, "success": true}})
		}
	}()

	jsonOK(w, map[string]string{"status": "downloading", "core": req.CoreName, "url": req.DownloadURL})
}

func (s *Server) handleCoreUpload(w http.ResponseWriter, r *http.Request) {
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

	destPath, err := s.coreSvc.Upload(coreName, header.Filename, file)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	jsonOK(w, map[string]string{"status": "uploaded", "core": coreName, "path": destPath})
}
