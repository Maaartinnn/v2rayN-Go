package core

import (
	"os/exec"
	"syscall"
)

// configureProcess 跨平台进程属性配置。
//
// Windows: 隐藏子进程控制台窗口，避免弹出黑框
func configureProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}

// killProcess 跨平台强杀内核进程。
//
// Windows: 直接杀单个进程（Process.Kill）
// 注：Windows 下完善的进程树绑定需使用 Job Object，在纯 Go 中实现较复杂，
// 此处使用 Process.Kill() 配合主程序的 Graceful Shutdown 钩子即可满足需求。
func killProcess(inst *coreInstance) {
	if inst.cmd.Process == nil {
		return
	}
	inst.cmd.Process.Kill()
}
