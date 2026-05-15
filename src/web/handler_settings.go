package web

import (
	"encoding/json"
	"net/http"

	"v2rayn-go/config"
	"v2rayn-go/service"
)

// SettingsHandler 设置管理独立处理器
type SettingsHandler struct {
	settingsSvc *service.SettingsService
	cfg         *config.AppConfig
}

// NewSettingsHandler 创建设置管理处理器
func NewSettingsHandler(settingsSvc *service.SettingsService, cfg *config.AppConfig) *SettingsHandler {
	return &SettingsHandler{
		settingsSvc: settingsSvc,
		cfg:         cfg,
	}
}

// Register 挂载设置管理路由
func (h *SettingsHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET  /api/settings{$}", h.handleGetSettings)
	mux.HandleFunc("POST /api/settings{$}", h.handleSaveSettings)

	mux.HandleFunc("GET  /api/proxy/system", h.handleGetSystemProxy)
	mux.HandleFunc("POST /api/proxy/system", h.handleSetSystemProxy)
}

func (h *SettingsHandler) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	settings := h.settingsSvc.GetSettings()
	jsonOK(w, settings)
}

func (h *SettingsHandler) handleSaveSettings(w http.ResponseWriter, r *http.Request) {
	var req service.UpdateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if err := h.settingsSvc.UpdateSettings(&req); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]string{"status": "saved"})
}

func (h *SettingsHandler) handleGetSystemProxy(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, map[string]interface{}{"enabled": false, "port": h.cfg.SocksPort})
}

func (h *SettingsHandler) handleSetSystemProxy(w http.ResponseWriter, r *http.Request) {
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
