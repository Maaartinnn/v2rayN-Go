//go:build !windows

package core

import (
	"os/exec"
	"syscall"
)

// configureProcess 跨平台进程属性配置。
//
// Linux/macOS: 设置独立进程组（Setpgid），当主进程遭遇 kill -9 时，
//
//	可通过 kill(-pid, SIGKILL) 杀掉整个进程组，防止内核进程变成孤儿进程
func configureProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// killProcess 跨平台强杀内核进程。
//
// Linux/macOS: 杀整个进程组（-pid），确保所有子线程一起终止
func killProcess(inst *coreInstance) {
	if inst.cmd.Process == nil {
		return
	}
	// 负数 PID 表示杀进程组（Setpgid 后的进程组 ID 等于进程 PID）
	syscall.Kill(-inst.cmd.Process.Pid, syscall.SIGKILL)
}
