package web

import (
	"encoding/json"
	"net/http"

	"v2rayn-go/database"
)

// RegisterStrategyGroupRoutes 注册策略组管理相关路由
func (s *Server) RegisterStrategyGroupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET    /api/strategy-groups/{$}", s.handleStrategyGroups)
	mux.HandleFunc("POST   /api/strategy-groups/{$}", s.handleStrategyGroupsCreate)

	mux.HandleFunc("GET    /api/strategy-groups/{uuid}", s.handleGetStrategyGroup)
	mux.HandleFunc("PUT    /api/strategy-groups/{uuid}", s.handleUpdateStrategyGroup)
	mux.HandleFunc("DELETE /api/strategy-groups/{uuid}", s.handleDeleteStrategyGroup)
}

func (s *Server) handleStrategyGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := s.strategySvc.List()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, groups)
}

func (s *Server) handleStrategyGroupsCreate(w http.ResponseWriter, r *http.Request) {
	var group database.StrategyGroup
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if err := s.strategySvc.Create(&group); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, group)
}

func (s *Server) handleGetStrategyGroup(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	group, err := s.strategySvc.Get(uuid)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonOK(w, group)
}

func (s *Server) handleUpdateStrategyGroup(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	var updated database.StrategyGroup
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	result, err := s.strategySvc.Update(uuid, &updated)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonOK(w, result)
}

func (s *Server) handleDeleteStrategyGroup(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if err := s.strategySvc.Delete(uuid); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]string{"status": "deleted"})
}
