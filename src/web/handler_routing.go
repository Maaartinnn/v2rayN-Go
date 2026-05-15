package web

import (
	"encoding/json"
	"net/http"

	"v2rayn-go/database"
)

// RegisterRoutingRuleRoutes 注册路由规则管理相关路由
func (s *Server) RegisterRoutingRuleRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET    /api/routing-rules/{$}", s.handleRoutingRules)
	mux.HandleFunc("POST   /api/routing-rules/{$}", s.handleRoutingRulesCreate)

	mux.HandleFunc("PUT    /api/routing-rules/{uuid}", s.handleUpdateRoutingRule)
	mux.HandleFunc("DELETE /api/routing-rules/{uuid}", s.handleDeleteRoutingRule)
}

func (s *Server) handleRoutingRules(w http.ResponseWriter, r *http.Request) {
	rules, err := s.routingSvc.List()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, rules)
}

func (s *Server) handleRoutingRulesCreate(w http.ResponseWriter, r *http.Request) {
	var rule database.RoutingRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if err := s.routingSvc.Create(&rule); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, rule)
}

func (s *Server) handleUpdateRoutingRule(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	var updated database.RoutingRule
	if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
		jsonError(w, "Invalid request", http.StatusBadRequest)
		return
	}
	result, err := s.routingSvc.Update(uuid, &updated)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonOK(w, result)
}

func (s *Server) handleDeleteRoutingRule(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	if err := s.routingSvc.Delete(uuid); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]string{"status": "deleted"})
}
