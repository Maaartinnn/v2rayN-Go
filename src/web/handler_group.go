package web

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"v2rayn-go/database"
	"v2rayn-go/subscription"
)

// RegisterGroupRoutes 注册分组管理相关路由
func (s *Server) RegisterGroupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET    /api/groups/{$}", s.handleGroups)
	mux.HandleFunc("POST   /api/groups/{$}", s.handleGroupsCreate)
	mux.HandleFunc("PUT    /api/groups/reorder", s.handleGroupsReorder)

	mux.HandleFunc("GET    /api/groups/{uuid}", s.handleGetGroup)
	mux.HandleFunc("PUT    /api/groups/{uuid}", s.handleUpdateGroup)
	mux.HandleFunc("DELETE /api/groups/{uuid}", s.handleDeleteGroup)
	mux.HandleFunc("POST   /api/groups/{uuid}/refresh", s.handleRefreshGroup)
	mux.HandleFunc("POST   /api/groups/{uuid}/refresh-proxy", s.handleRefreshGroupProxy)
}

func (s *Server) handleGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := s.groupSvc.List()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, groups)
}

func (s *Server) handleGroupsCreate(w http.ResponseWriter, r *http.Request) {
	var group database.NodeGroup
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if err := s.groupSvc.Create(&group); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, group)
}

func (s *Server) handleGroupsReorder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UUID       string `json:"uuid"`
		BeforeUUID string `json:"before_uuid"`
		AfterUUID  string `json:"after_uuid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	newOrder, err := s.groupSvc.Reorder(req.UUID, req.BeforeUUID, req.AfterUUID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}

	jsonOK(w, map[string]interface{}{"status": "reordered", "sort_order": newOrder})
}

func (s *Server) handleGetGroup(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	group, err := s.groupSvc.Get(uuid)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonOK(w, group)
}

func (s *Server) handleUpdateGroup(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	var updated database.NodeGroup
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	result, err := s.groupSvc.Update(uuid, &updated)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, result)
}

func (s *Server) handleDeleteGroup(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if err := s.groupSvc.Delete(uuid); err != nil {
		if strings.Contains(err.Error(), "last group") || strings.Contains(err.Error(), "not found") {
			jsonError(w, err.Error(), http.StatusBadRequest)
		} else {
			jsonError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	jsonOK(w, map[string]string{"status": "deleted"})
}

func (s *Server) handleRefreshGroup(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	group, err := s.groupSvc.Get(uuid)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
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

func (s *Server) handleRefreshGroupProxy(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	group, err := s.groupSvc.Get(uuid)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
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
