package web

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"v2rayn-go/core"
	"v2rayn-go/service"

	"github.com/gorilla/websocket"
)

// StatusBroadcaster WebSocket 广播接口，供其他 Handler 注入使用
type StatusBroadcaster interface {
	BroadcastStatus()
	Broadcast(msg interface{})
}

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

// WSHandler WebSocket 独立处理器
type WSHandler struct {
	coreSvc  *service.CoreService
	coreMgr  *core.CoreAdminManager
	upgrader websocket.Upgrader
	clients  sync.Map
}

// NewWSHandler 创建 WebSocket 处理器
func NewWSHandler(coreSvc *service.CoreService, coreMgr *core.CoreAdminManager) *WSHandler {
	return &WSHandler{
		coreSvc: coreSvc,
		coreMgr: coreMgr,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// Register 挂载 WebSocket 路由
func (h *WSHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/ws", h.handleWebSocket)
}

func (h *WSHandler) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	wc := &wsConn{conn: conn}
	defer wc.Close()

	clientID := fmt.Sprintf("%p", conn)
	h.clients.Store(clientID, wc)
	defer h.clients.Delete(clientID)

	h.sendStatusToClient(wc)

	for {
		_, _, err := wc.ReadMessage()
		if err != nil {
			break
		}
	}
}

// ========== StatusBroadcaster 接口实现 ==========

// BroadcastStatus 广播核心状态（实现 StatusBroadcaster 接口）
func (h *WSHandler) BroadcastStatus() {
	statuses := h.coreSvc.GetAllStatus()
	h.Broadcast(map[string]interface{}{"type": "status", "payload": statuses})
}

// Broadcast 广播任意消息（实现 StatusBroadcaster 接口）
func (h *WSHandler) Broadcast(msg interface{}) {
	h.clients.Range(func(key, value interface{}) bool {
		if wc, ok := value.(*wsConn); ok {
			if err := wc.WriteJSON(msg); err != nil {
				h.clients.Delete(key)
			}
		}
		return true
	})
}

func (h *WSHandler) sendStatusToClient(wc *wsConn) {
	statuses := h.coreSvc.GetAllStatus()
	wc.WriteJSON(map[string]interface{}{"type": "status", "payload": statuses})
}

// LogBroadcaster 启动日志广播 goroutine（由 Server.Start 调用）
func (h *WSHandler) LogBroadcaster() {
	logChan := h.coreMgr.LogChannel()
	for entry := range logChan {
		h.Broadcast(map[string]interface{}{"type": "log", "payload": entry})
	}
}
