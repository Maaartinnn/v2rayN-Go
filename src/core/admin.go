package core

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"v2rayn-go/config"
)

// CoreType 内核类型
type CoreType string

const (
	CoreTypeXray    CoreType = "xray"
	CoreTypeSingBox CoreType = "sing-box"
)

// CoreStatus 内核运行状态
type CoreStatus string

const (
	StatusStopped  CoreStatus = "stopped"
	StatusRunning  CoreStatus = "running"
	StatusStarting CoreStatus = "starting"
	StatusError    CoreStatus = "error"
)

// CoreInfo 内核运行时信息
type CoreInfo struct {
	Type      CoreType   `json:"type"`
	Status    CoreStatus `json:"status"`
	PID       int        `json:"pid"`
	StartTime time.Time  `json:"start_time"`
	ErrorMsg  string     `json:"error_msg"`
}

// LogEntry 日志条目
type LogEntry struct {
	Time    time.Time `json:"time"`
	Level   string    `json:"level"`
	Content string    `json:"content"`
}

// CoreAdminManager 管理外部内核进程的生命周期
type CoreAdminManager struct {
	cfg *config.AppConfig

	mu      sync.RWMutex
	cores   map[CoreType]*coreInstance
	logChan chan LogEntry
}

type coreInstance struct {
	info    CoreInfo
	cmd     *exec.Cmd
	cancel  context.CancelFunc
	logFile *os.File
}

// NewCoreAdminManager 创建新的内核管理器
func NewCoreAdminManager(cfg *config.AppConfig) *CoreAdminManager {
	return &CoreAdminManager{
		cfg:     cfg,
		cores:   make(map[CoreType]*coreInstance),
		logChan: make(chan LogEntry, 100),
	}
}

// StartCore 启动指定类型的内核
func (m *CoreAdminManager) StartCore(coreType CoreType, configPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已在运行
	if inst, ok := m.cores[coreType]; ok && inst.info.Status == StatusRunning {
		return fmt.Errorf("core %s is already running (PID: %d)", coreType, inst.info.PID)
	}

	// 获取内核可执行文件路径
	binPath := m.getCoreBinaryPath(coreType)
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		return fmt.Errorf("core binary not found: %s", binPath)
	}

	// 检查配置文件
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", configPath)
	}

	// 创建 context 用于控制进程生命周期
	ctx, cancel := context.WithCancel(context.Background())

	// 构建命令参数
	args := m.buildCoreArgs(coreType, configPath)
	cmd := exec.CommandContext(ctx, binPath, args...)
	cmd.Dir = filepath.Dir(configPath)

	// 设置 stdout/stderr 管道
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// 创建日志文件
	logPath := filepath.Join(m.cfg.LogDir, fmt.Sprintf("%s.log", coreType))
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create log file: %w", err)
	}

	// 更新实例信息
	inst := &coreInstance{
		info: CoreInfo{
			Type:      coreType,
			Status:    StatusStarting,
			StartTime: time.Now(),
		},
		cmd:     cmd,
		cancel:  cancel,
		logFile: logFile,
	}
	m.cores[coreType] = inst

	// 启动进程
	if err := cmd.Start(); err != nil {
		cancel()
		logFile.Close()
		inst.info.Status = StatusError
		inst.info.ErrorMsg = err.Error()
		return fmt.Errorf("failed to start core: %w", err)
	}

	inst.info.PID = cmd.Process.Pid
	inst.info.Status = StatusRunning

	// 启动日志收集 goroutine
	go m.collectLogs(coreType, stdoutPipe, logFile, "stdout")
	go m.collectLogs(coreType, stderrPipe, logFile, "stderr")

	// 启动进程监控 goroutine
	go func() {
		err := cmd.Wait()
		m.mu.Lock()
		defer m.mu.Unlock()

		if inst, ok := m.cores[coreType]; ok {
			if err != nil {
				inst.info.Status = StatusError
				inst.info.ErrorMsg = err.Error()
				m.emitLog(coreType, "error", fmt.Sprintf("Core exited with error: %v", err))
			} else {
				inst.info.Status = StatusStopped
				m.emitLog(coreType, "info", "Core exited normally")
			}
			if inst.logFile != nil {
				inst.logFile.Close()
			}
		}
	}()

	m.emitLog(coreType, "info", fmt.Sprintf("Core started (PID: %d)", cmd.Process.Pid))
	return nil
}

// StopCore 停止指定类型的内核
func (m *CoreAdminManager) StopCore(coreType CoreType) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	inst, ok := m.cores[coreType]
	if !ok || inst.info.Status != StatusRunning {
		return fmt.Errorf("core %s is not running", coreType)
	}

	// 通过 cancel 发送优雅关闭信号
	inst.cancel()

	// 等待进程退出（最多 5 秒）
	done := make(chan struct{})
	go func() {
		if inst.cmd.Process != nil {
			inst.cmd.Process.Wait()
		}
		close(done)
	}()

	select {
	case <-done:
		m.emitLog(coreType, "info", "Core stopped gracefully")
	case <-time.After(5 * time.Second):
		// 超时后强制杀死进程
		if inst.cmd.Process != nil {
			inst.cmd.Process.Kill()
		}
		m.emitLog(coreType, "warn", "Core force killed after timeout")
	}

	inst.info.Status = StatusStopped
	return nil
}

// StopAll 停止所有运行中的内核
func (m *CoreAdminManager) StopAll() {
	m.mu.RLock()
	types := make([]CoreType, 0, len(m.cores))
	for t, inst := range m.cores {
		if inst.info.Status == StatusRunning {
			types = append(types, t)
		}
	}
	m.mu.RUnlock()

	for _, t := range types {
		if err := m.StopCore(t); err != nil {
			log.Printf("Failed to stop core %s: %v", t, err)
		}
	}
}

// GetCoreStatus 获取指定内核的状态
func (m *CoreAdminManager) GetCoreStatus(coreType CoreType) CoreInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if inst, ok := m.cores[coreType]; ok {
		return inst.info
	}
	return CoreInfo{Type: coreType, Status: StatusStopped}
}

// GetAllStatus 获取所有内核状态
func (m *CoreAdminManager) GetAllStatus() []CoreInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statuses := make([]CoreInfo, 0, len(m.cores))
	for _, inst := range m.cores {
		statuses = append(statuses, inst.info)
	}
	return statuses
}

// LogChannel 返回日志通道（供 WebSocket 使用）
func (m *CoreAdminManager) LogChannel() <-chan LogEntry {
	return m.logChan
}

// getCoreBinaryPath 获取内核可执行文件路径
func (m *CoreAdminManager) getCoreBinaryPath(coreType CoreType) string {
	binName := string(coreType)
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	return filepath.Join(m.cfg.BinDir, binName)
}

// buildCoreArgs 构建内核启动参数
func (m *CoreAdminManager) buildCoreArgs(coreType CoreType, configPath string) []string {
	switch coreType {
	case CoreTypeXray:
		return []string{"run", "-config", configPath}
	case CoreTypeSingBox:
		return []string{"run", "-c", configPath}
	default:
		return []string{"run", "-config", configPath}
	}
}

// collectLogs 从进程管道收集日志
func (m *CoreAdminManager) collectLogs(coreType CoreType, reader io.Reader, logFile *os.File, source string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if logFile != nil {
			fmt.Fprintf(logFile, "[%s] %s\n", source, line)
		}
		m.emitLog(coreType, "info", line)
	}
}

// emitLog 发送日志到通道
func (m *CoreAdminManager) emitLog(coreType CoreType, level, content string) {
	entry := LogEntry{
		Time:    time.Now(),
		Level:   level,
		Content: fmt.Sprintf("[%s] %s", coreType, content),
	}

	select {
	case m.logChan <- entry:
	default:
		// 通道满时丢弃旧日志
		select {
		case <-m.logChan:
		default:
		}
		m.logChan <- entry
	}
}
