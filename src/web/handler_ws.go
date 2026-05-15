package web

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// wsConn 封装 WebSocket 连接，加写锁防止并发写入 panic
type wsConn struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (w *wsConn) WriteJSON(v interface{}) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn.WriteJSON(v)
}

func (w *wsConn) ReadMessage() (messageType int, p []byte, err error) {
	return w.conn.ReadMessage()
}

func (w *wsConn) Close() error {
	return w.conn.Close()
}

// RegisterWebSocketRoutes 注册 WebSocket 路由
func (s *Server) RegisterWebSocketRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/ws", s.handleWebSocket)
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	wc := &wsConn{conn: conn}
	defer wc.Close()

	clientID := fmt.Sprintf("%p", conn)
	s.wsClients.Store(clientID, wc)
	defer s.wsClients.Delete(clientID)

	s.sendStatusToClient(wc)

	for {
		_, _, err := wc.ReadMessage()
		if err != nil {
			break
		}
	}
}

// ========== WebSocket Broadcasting ==========

func (s *Server) broadcastToAll(msg interface{}) {
	s.wsClients.Range(func(key, value interface{}) bool {
		if wc, ok := value.(*wsConn); ok {
			if err := wc.WriteJSON(msg); err != nil {
				s.wsClients.Delete(key)
			}
		}
		return true
	})
}

func (s *Server) broadcastStatus() {
	statuses := s.coreSvc.GetAllStatus()
	s.broadcastToAll(map[string]interface{}{"type": "status", "payload": statuses})
}

func (s *Server) sendStatusToClient(wc *wsConn) {
	statuses := s.coreSvc.GetAllStatus()
	wc.WriteJSON(map[string]interface{}{"type": "status", "payload": statuses})
}

func (s *Server) logBroadcaster() {
	logChan := s.coreMgr.LogChannel()
	for entry := range logChan {
		s.broadcastToAll(map[string]interface{}{"type": "log", "payload": entry})
	}
}
