package core

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"v2rayn-go/config"
	"v2rayn-go/coredef"
)

// CoreType 内核类型（类型别名，指向 coredef.CoreType）
type CoreType = coredef.CoreType

// 内核类型常量（兼容引用，实际定义在 coredef 包中）
const (
	CoreTypeXray    = coredef.TypeXray
	CoreTypeSingBox = coredef.TypeSingBox
	CoreTypeMihomo  = coredef.TypeMihomo
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
	Source  string    `json:"source"` // 日志来源: "system", "xray", "sing-box", "mihomo"
}

// StartOption 启动选项（Functional Options 模式）
type StartOption func(*startConfig)

// startConfig 内部启动配置
type startConfig struct {
	configData []byte // 非空时使用 stdin 模式（无文件落地）
}

// WithStdin 使用 stdin 注入配置（无文件落地模式）。
// 配置数据直接通过 cmd.Stdin 管道传给内核进程，全程不触碰物理磁盘。
func WithStdin(data []byte) StartOption {
	return func(sc *startConfig) {
		sc.configData = data
	}
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
	done    chan struct{} // 监控 goroutine 完成后关闭，用于 StopCore 解锁等待
}

// NewCoreAdminManager 创建新的内核管理器
func NewCoreAdminManager(cfg *config.AppConfig) *CoreAdminManager {
	return &CoreAdminManager{
		cfg:     cfg,
		cores:   make(map[CoreType]*coreInstance),
		logChan: make(chan LogEntry, coredef.CoreLogChannelBuffer),
	}
}

// StartCore 启动指定类型的内核。
//
// 支持两种启动模式：
//   - 文件模式（默认）：传入 configPath，内核从文件读取配置
//   - stdin 模式：通过 WithStdin(data) 选项注入配置，全程无文件落地
//
// 示例：
//
//	// 文件模式（向后兼容）
//	m.StartCore(coreType, "/path/to/config.json")
//
//	// stdin 模式（无文件落地）
//	m.StartCore(coreType, "", core.WithStdin(configData))
func (m *CoreAdminManager) StartCore(coreType CoreType, configPath string, opts ...StartOption) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已在运行
	if inst, ok := m.cores[coreType]; ok && inst.info.Status == StatusRunning {
		return fmt.Errorf("core %s is already running (PID: %d)", coreType, inst.info.PID)
	}

	// 获取内核可执行文件路径
	binPath := m.getCoreBinaryPath(coreType)
	if _, err := os.Stat(binPath); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("core binary not found: %s", binPath)
	}

	// 解析启动选项
	sc := &startConfig{}
	for _, opt := range opts {
		opt(sc)
	}

	useStdin := sc.configData != nil

	// 文件模式下检查配置文件是否存在
	if !useStdin {
		if _, err := os.Stat(configPath); errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("config file not found: %s", configPath)
		}
	}

	// 创建 context 用于控制进程生命周期
	ctx, cancel := context.WithCancel(context.Background())

	// 构建命令参数和工作目录
	var args []string
	var workDir string
	if useStdin {
		args = m.buildStdinArgs(coreType)
		// stdin 模式下工作目录设为内核二进制所在目录（Mihomo 需要在此加载 geoip.db 等资源）
		workDir = filepath.Join(m.cfg.BinDir, coredef.Registry[coreType].SubDir)
	} else {
		args = m.buildCoreArgs(coreType, configPath)
		workDir = filepath.Dir(configPath)
	}

	cmd := exec.CommandContext(ctx, binPath, args...)
	cmd.Dir = workDir

	// stdin 模式：将配置数据绑定到子进程的 Stdin
	if useStdin {
		cmd.Stdin = bytes.NewReader(sc.configData)
	}

	// 跨平台进程属性配置
	configureProcess(cmd)

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
		done:    make(chan struct{}), // 监控 goroutine 结束时关闭此通道
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

	// 启动进程监控 goroutine（唯一的 cmd.Wait 调用点和 logFile.Close 点）
	go func() {
		err := cmd.Wait() // 阻塞直到进程退出；StopCore 不再另开 Wait

		m.mu.Lock()
		defer m.mu.Unlock()
		defer close(inst.done) // 通知 StopCore：清理完成，可以安全返回

		if inst, ok := m.cores[coreType]; ok {
			// 只有监控 goroutine 更新进程退出状态，避免与 StopCore 竞争写入
			if err != nil {
				inst.info.Status = StatusError
				inst.info.ErrorMsg = err.Error()
				m.emitLog(coreType, "error", fmt.Sprintf("Core exited with error: %v", err))
			} else {
				inst.info.Status = StatusStopped
				m.emitLog(coreType, "info", "Core exited normally")
			}
			// logFile 只在此处关闭，StopCore 不再重复关闭
			if inst.logFile != nil {
				inst.logFile.Close()
			}
		}
	}()

	mode := "file"
	if useStdin {
		mode = "stdin"
	}
	m.emitLog(coreType, "info", fmt.Sprintf("Core started (PID: %d, mode: %s)", cmd.Process.Pid, mode))
	return nil
}

// StopCore 停止指定类型的内核
// 设计要点：
//   - 锁只用于读取实例引用和发送取消信号，不持有锁等待进程退出，避免死锁
//   - 不调用 cmd.Wait() 或 Process.Wait()，由 StartCore 的监控 goroutine 唯一负责
//   - 通过 inst.done 通道等待监控 goroutine 完成清理（状态更新 + logFile.Close）
func (m *CoreAdminManager) StopCore(coreType CoreType) error {
	m.mu.Lock()
	inst, ok := m.cores[coreType]
	if !ok || inst.info.Status != StatusRunning {
		m.mu.Unlock()
		return fmt.Errorf("core %s is not running", coreType)
	}
	// 保存引用，立即释放锁，避免阻塞监控 goroutine
	inst.cancel()
	done := inst.done
	m.mu.Unlock()

	// 在锁外等待进程退出（监控 goroutine 完成 cmd.Wait 后关闭 done 通道）
	select {
	case <-done:
		// 监控 goroutine 已完成状态更新和 logFile.Close
		m.emitLog(coreType, "info", "Core stopped gracefully")
	case <-time.After(coredef.CoreStopTimeout):
		// 超时后强制杀死进程（跨平台：Unix 杀进程组，Windows 杀单进程）
		killProcess(inst)
		m.emitLog(coreType, "warn", "Core force killed after timeout")
		<-done // 等待监控 goroutine 完成最终清理
	}

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
			slog.Error("failed to stop core", "core", string(t), "error", err)
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

// getCoreBinaryPath 获取内核可执行文件路径 (bin/xray/xray.exe, bin/sing_box/sing-box.exe, bin/mihomo/mihomo.exe)
func (m *CoreAdminManager) getCoreBinaryPath(coreType CoreType) string {
	meta := coredef.Registry[coreType]
	return filepath.Join(m.cfg.BinDir, meta.SubDir, meta.BinaryName())
}

// buildCoreArgs 构建内核启动参数（文件模式）
func (m *CoreAdminManager) buildCoreArgs(coreType CoreType, configPath string) []string {
	switch coreType {
	case CoreTypeXray:
		return []string{"run", "-config", configPath}
	case CoreTypeSingBox:
		return []string{"run", "-c", configPath}
	case CoreTypeMihomo:
		return []string{"-f", configPath}
	default:
		return []string{"run", "-config", configPath}
	}
}

// buildStdinArgs 构建内核启动参数（stdin 无文件落地模式）。
//
// 各内核对 stdin 的支持：
//   - Xray:      -config stdin:     （官方原生支持）
//   - Sing-box:  -c stdin:          （官方原生支持）
//   - Mihomo:    -f -               （用 - 代替文件路径）
//
// 注意：Mihomo 的 -d . 依赖外部 cmd.Dir 已被设置为内核工作目录
func (m *CoreAdminManager) buildStdinArgs(coreType CoreType) []string {
	switch coreType {
	case CoreTypeXray:
		return []string{"run", "-config", "stdin:"}
	case CoreTypeSingBox:
		return []string{"run", "-c", "stdin:"}
	case CoreTypeMihomo:
		return []string{"-d", ".", "-f", "-"}
	default:
		return []string{"run", "-config", "stdin:"}
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
	if err := scanner.Err(); err != nil {
		m.emitLog(coreType, "error", fmt.Sprintf("Log scanner error: %v", err))
	}
}

// emitLog 发送日志到通道
func (m *CoreAdminManager) emitLog(coreType CoreType, level, content string) {
	entry := LogEntry{
		Time:    time.Now(),
		Level:   level,
		Content: content,
		Source:  string(coreType),
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
