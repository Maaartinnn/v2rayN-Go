package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"v2rayn-go/core"
	"v2rayn-go/service"

	"github.com/gorilla/websocket"
)

// ========== 状态广播接口 ==========

// StatusBroadcaster WebSocket 广播接口，供其他 Handler 注入使用
type StatusBroadcaster interface {
	BroadcastStatus()
	Broadcast(msg any)
}

// ========== 泛型环形缓冲区 ==========

const (
	ringBufferSize = 500 // 日志缓冲区容量
	sendChanSize   = 50  // 每个连接的发送缓冲区容量
	pingInterval   = 30  // Ping 间隔（秒）
	pongWait       = 60  // Pong 等待超时（秒）
	writeWait      = 10  // 写操作超时（秒）
)

// ringBuffer 并发安全的泛型环形缓冲区。
// 写入时覆盖最旧的数据，Snapshot 返回深拷贝以避免读写竞争。
type ringBuffer[T any] struct {
	mu     sync.RWMutex
	data   []T
	size   int
	cursor int // 下一个写入位置
	full   bool
}

// newRingBuffer 创建指定容量的环形缓冲区
func newRingBuffer[T any](size int) *ringBuffer[T] {
	return &ringBuffer[T]{
		data: make([]T, size),
		size: size,
	}
}

// Add 向缓冲区追加一条记录，超出容量时覆盖最旧的
func (r *ringBuffer[T]) Add(item T) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.data[r.cursor] = item
	r.cursor = (r.cursor + 1) % r.size
	if r.cursor == 0 {
		r.full = true
	}
}

// Snapshot 按时间顺序返回缓冲区内容的深拷贝切片。
// 返回全新切片，与内部 data 无共享引用，json.Marshal 不会与 Add 产生竞态。
func (r *ringBuffer[T]) Snapshot() []T {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.full {
		// 缓冲区未满：有效数据在 [0, cursor)
		res := make([]T, r.cursor)
		copy(res, r.data[:r.cursor])
		return res
	}

	// 缓冲区已满：cursor 指向最旧数据的起始位置
	res := make([]T, r.size)
	n := copy(res, r.data[r.cursor:]) // 从 cursor 到末尾（较旧）
	copy(res[n:], r.data[:r.cursor])  // 从头到 cursor（较新）
	return res
}

// ========== WebSocket 连接封装 ==========

// wsConn 封装单个 WebSocket 连接。
// 所有写操作都通过 sendCh 串行化到 writePump goroutine，实现无锁并发安全。
type wsConn struct {
	conn      *websocket.Conn
	sendCh    chan any      // 业务消息缓冲通道
	done      chan struct{} // 关闭信号
	closeOnce sync.Once     // 保证关闭操作幂等执行
}

// close 幂等关闭连接。
// 由三个触发源（Read 超时、writePump 写失败、Broadcast 慢客户端踢出）共同调用，
// sync.Once 确保 close(done) 和 conn.Close() 只执行一次，避免 panic。
func (w *wsConn) close() {
	w.closeOnce.Do(func() {
		close(w.done)
		w.conn.Close()
	})
}

// writePump 消费 sendCh 并定时发送 Ping。
// 这是该连接唯一的写 goroutine，所有 WriteMessage / WriteControl 均在此串行执行，
// 因此无需额外互斥锁来保护并发写。
func (w *wsConn) writePump() {
	ticker := time.NewTicker(pingInterval * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-w.sendCh:
			// sendCh 已关闭（连接已从 clients 中移除）
			if !ok {
				return
			}
			w.conn.SetWriteDeadline(time.Now().Add(writeWait * time.Second))
			if err := w.conn.WriteJSON(msg); err != nil {
				return // 写失败，退出 pump，Read 循环会清理连接
			}

		case <-ticker.C:
			// 周期性 Ping，保活连接并检测死连接
			w.conn.SetWriteDeadline(time.Now().Add(writeWait * time.Second))
			if err := w.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait*time.Second)); err != nil {
				return // Ping 失败，连接已断开
			}

		case <-w.done:
			return // 连接已被外部关闭
		}
	}
}

// ========== WebSocket 处理器 ==========

// WSHandler WebSocket 独立处理器
type WSHandler struct {
	coreSvc   *service.CoreService
	coreMgr   *core.CoreAdminManager
	upgrader  websocket.Upgrader
	clients   sync.Map                   // map[string]*wsConn
	logBuffer *ringBuffer[core.LogEntry] // 最近日志环形缓冲
}

// NewWSHandler 创建 WebSocket 处理器
func NewWSHandler(coreSvc *service.CoreService, coreMgr *core.CoreAdminManager) *WSHandler {
	return &WSHandler{
		coreSvc:   coreSvc,
		coreMgr:   coreMgr,
		logBuffer: newRingBuffer[core.LogEntry](ringBufferSize),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// Register 挂载 WebSocket 路由
func (h *WSHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/ws", h.handleWebSocket)
}

// handleWebSocket 处理 WebSocket 升级与连接生命周期
func (h *WSHandler) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// 初始化连接对象，sendCh 缓冲区满时发送方会被丢弃（非阻塞）
	wc := &wsConn{
		conn:   conn,
		sendCh: make(chan any, sendChanSize),
		done:   make(chan struct{}),
	}

	clientID := fmt.Sprintf("%p", conn)
	h.clients.Store(clientID, wc)

	// 设置 Pong 处理器：收到客户端的 Pong 响应时续期 ReadDeadline
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait * time.Second))
		return nil
	})

	// 先启动写泵，再投递初始消息，保证所有写操作都经过 writePump
	go wc.writePump()

	// 通过 sendCh 投递初始状态（核心运行状态 + 最近日志快照）
	statuses := h.coreSvc.GetAllStatus()
	wc.sendCh <- map[string]any{"type": "status", "payload": statuses}
	wc.sendCh <- map[string]any{"type": "log_snapshot", "payload": h.logBuffer.Snapshot()}

	// Read 循环：设置初始 ReadDeadline，等待客户端消息或超时
	conn.SetReadDeadline(time.Now().Add(pongWait * time.Second))
	for {
		_, _, err := wc.conn.ReadMessage()
		if err != nil {
			// 客户端断开、ReadDeadline 超时、或 writePump 先退出导致连接关闭
			h.removeClient(clientID, wc)
			return
		}
		// 客户端发送了消息（当前协议无业务消息，仅用于检测连接存活）
	}
}

// removeClient 从 clients 中移除连接并幂等关闭。
// 可能由 Read 循环、Broadcast（慢客户端）、writePump 同时触发，
// sync.Once + close(wc.done) 确保安全。
func (h *WSHandler) removeClient(clientID string, wc *wsConn) {
	h.clients.Delete(clientID)
	wc.close()
}

// ========== StatusBroadcaster 接口实现 ==========

// BroadcastStatus 广播核心状态（实现 StatusBroadcaster 接口）
func (h *WSHandler) BroadcastStatus() {
	statuses := h.coreSvc.GetAllStatus()
	h.Broadcast(map[string]any{"type": "status", "payload": statuses})
}

// Broadcast 向所有连接非阻塞投递消息。
// sendCh 满时认为该客户端为慢客户端，主动踢出以保护其他连接不受影响。
func (h *WSHandler) Broadcast(msg any) {
	h.clients.Range(func(key, value any) bool {
		wc := value.(*wsConn)
		select {
		case wc.sendCh <- msg:
			// 投递成功
		default:
			// 缓冲区满，该客户端处理过慢，主动断开
			h.removeClient(key.(string), wc)
		}
		return true
	})
}

// ========== 日志广播 ==========

// LogBroadcaster 启动日志广播 goroutine（由 Server.Start 在独立 goroutine 中调用）。
// 从 CoreAdminManager 的日志通道接收日志，写入环形缓冲区并广播给所有客户端。
// 支持通过 context 取消以实现优雅退出。
func (h *WSHandler) LogBroadcaster(ctx context.Context) {
	logChan := h.coreMgr.LogChannel()
	for {
		select {
		case <-ctx.Done():
			return // 服务端关闭，优雅退出
		case entry, ok := <-logChan:
			if !ok {
				return // 日志通道已关闭
			}
			h.logBuffer.Add(entry)
			h.Broadcast(map[string]any{"type": "log", "payload": entry})
		}
	}
}
