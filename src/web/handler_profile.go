package web

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"v2rayn-go/database"
	"v2rayn-go/parser"
)

// RegisterProfileRoutes 注册节点管理相关路由
func (s *Server) RegisterProfileRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET    /api/profiles/{$}", s.handleProfiles)
	mux.HandleFunc("POST   /api/profiles/{$}", s.handleProfilesCreate)
	mux.HandleFunc("POST   /api/profiles/import{$}", s.handleProfileImport)
	mux.HandleFunc("POST   /api/profiles/import-image", s.handleProfileImportImage)
	mux.HandleFunc("POST   /api/profiles/dedup", s.handleProfileDedup)
	mux.HandleFunc("POST   /api/profiles/ping-all", s.handlePingAll)

	mux.HandleFunc("GET    /api/profiles/{uuid}", s.handleGetProfile)
	mux.HandleFunc("PUT    /api/profiles/{uuid}", s.handleUpdateProfile)
	mux.HandleFunc("DELETE /api/profiles/{uuid}", s.handleDeleteProfile)
	mux.HandleFunc("POST   /api/profiles/{uuid}/select", s.handleSelectProfile)
	mux.HandleFunc("POST   /api/profiles/{uuid}/ping", s.handlePingProfile)
}

func (s *Server) handleProfiles(w http.ResponseWriter, r *http.Request) {
	profiles, err := s.profileSvc.List()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, profiles)
}

func (s *Server) handleProfilesCreate(w http.ResponseWriter, r *http.Request) {
	var profile database.Profile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if err := s.profileSvc.Create(&profile); err != nil {
		if strings.Contains(err.Error(), "not found") {
			jsonError(w, err.Error(), http.StatusBadRequest)
		} else {
			jsonError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	jsonOK(w, profile)
}

func (s *Server) handleProfileImport(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Links     string `json:"links"`
		GroupUUID string `json:"group_uuid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	count, err := s.profileSvc.ImportLinks(req.Links, req.GroupUUID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "required") {
			jsonError(w, err.Error(), http.StatusBadRequest)
		} else {
			jsonError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	jsonOK(w, map[string]int{"imported": count})
}

func (s *Server) handleProfileImportImage(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
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
		s.importParsedLinks(w, links, groupUUID)
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

	s.importParsedLinks(w, links, groupUUID)
}

// importParsedLinks 将已解析的链接导入到指定分组
func (s *Server) importParsedLinks(w http.ResponseWriter, links []string, groupUUID string) {
	count, err := s.profileSvc.ImportParsedLinks(links, groupUUID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "required") {
			jsonError(w, err.Error(), http.StatusBadRequest)
		} else {
			jsonError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	jsonOK(w, map[string]int{"imported": count})
}

func (s *Server) handleProfileDedup(w http.ResponseWriter, r *http.Request) {
	var req struct {
		GroupUUID string `json:"group_uuid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	result, err := s.profileSvc.Dedup(req.GroupUUID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonOK(w, map[string]interface{}{
		"removed": result.Removed,
		"total":   result.Total,
	})
}

func (s *Server) handlePingAll(w http.ResponseWriter, r *http.Request) {
	go s.pingSvc.PingAllProfiles(r.Context(), 20)
	jsonOK(w, map[string]string{"status": "pinging"})
}

func (s *Server) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	profile, err := s.profileSvc.Get(uuid)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonOK(w, profile)
}

func (s *Server) handleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	profile, err := s.profileSvc.Update(uuid, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "required") {
			jsonError(w, err.Error(), http.StatusBadRequest)
		} else {
			jsonError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	jsonOK(w, profile)
}

func (s *Server) handleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if err := s.profileSvc.Delete(uuid); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]string{"status": "deleted"})
}

func (s *Server) handleSelectProfile(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if err := s.profileSvc.Select(uuid); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]string{"status": "selected"})
}

func (s *Server) handlePingProfile(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	profile, err := s.profileSvc.Get(uuid)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	go s.pingSvc.PingSingleProfile(profile)
	jsonOK(w, map[string]string{"status": "pinging"})
}
