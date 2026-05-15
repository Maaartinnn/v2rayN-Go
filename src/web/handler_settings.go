package web

import (
	"encoding/json"
	"net/http"

	"v2rayn-go/service"
)

// RegisterSettingsRoutes 注册设置和系统代理相关路由
func (s *Server) RegisterSettingsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET  /api/settings{$}", s.handleGetSettings)
	mux.HandleFunc("POST /api/settings{$}", s.handleSaveSettings)

	mux.HandleFunc("GET  /api/proxy/system", s.handleGetSystemProxy)
	mux.HandleFunc("POST /api/proxy/system", s.handleSetSystemProxy)
}

func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	settings := s.settingsSvc.GetSettings()
	jsonOK(w, settings)
}

func (s *Server) handleSaveSettings(w http.ResponseWriter, r *http.Request) {
	var req service.UpdateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if err := s.settingsSvc.UpdateSettings(&req); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]string{"status": "saved"})
}

func (s *Server) handleGetSystemProxy(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, map[string]interface{}{"enabled": false, "port": s.cfg.SocksPort})
}

func (s *Server) handleSetSystemProxy(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Enabled bool `json:"enabled"`
		Port    int  `json:"port"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	jsonOK(w, map[string]interface{}{"enabled": req.Enabled, "port": req.Port})
}
