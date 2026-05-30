package core

import (
	"testing"

	"v2rayn-go/config"
)

// ========== buildStdinArgs 测试 ==========

func newTestManager(t *testing.T) *CoreAdminManager {
	cfg := &config.AppConfig{
		AppDir: t.TempDir(),
		BinDir: "/tmp/bin",
		LogDir: "/tmp/logs",
	}
	return NewCoreAdminManager(cfg)
}

func TestBuildStdinArgs_Xray(t *testing.T) {
	m := newTestManager(t)
	args := m.buildStdinArgs(CoreTypeXray)
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d: %v", len(args), args)
	}
	if args[0] != "run" || args[1] != "-config" || args[2] != "stdin:" {
		t.Fatalf("unexpected xray stdin args: %v", args)
	}
}

func TestBuildStdinArgs_SingBox(t *testing.T) {
	m := newTestManager(t)
	args := m.buildStdinArgs(CoreTypeSingBox)
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d: %v", len(args), args)
	}
	if args[0] != "run" || args[1] != "-c" || args[2] != "stdin:" {
		t.Fatalf("unexpected sing-box stdin args: %v", args)
	}
}

func TestBuildStdinArgs_Mihomo(t *testing.T) {
	m := newTestManager(t)
	args := m.buildStdinArgs(CoreTypeMihomo)
	if len(args) != 4 {
		t.Fatalf("expected 4 args, got %d: %v", len(args), args)
	}
	if args[0] != "-d" || args[1] != "." || args[2] != "-f" || args[3] != "-" {
		t.Fatalf("unexpected mihomo stdin args: %v", args)
	}
}

func TestBuildStdinArgs_Unknown(t *testing.T) {
	m := newTestManager(t)
	args := m.buildStdinArgs(CoreType("unknown"))
	if len(args) != 3 {
		t.Fatalf("expected 3 args for unknown type, got %d: %v", len(args), args)
	}
	if args[2] != "stdin:" {
		t.Fatalf("expected stdin: for unknown type, got %v", args[2])
	}
}

// ========== buildCoreArgs 测试（文件模式）==========

func TestBuildCoreArgs_Xray(t *testing.T) {
	m := newTestManager(t)
	args := m.buildCoreArgs(CoreTypeXray, "/path/to/config.json")
	if args[2] != "/path/to/config.json" {
		t.Fatalf("expected config path, got %v", args[2])
	}
}

func TestBuildCoreArgs_SingBox(t *testing.T) {
	m := newTestManager(t)
	args := m.buildCoreArgs(CoreTypeSingBox, "/path/to/config.json")
	if args[1] != "-c" || args[2] != "/path/to/config.json" {
		t.Fatalf("unexpected sing-box file args: %v", args)
	}
}

func TestBuildCoreArgs_Mihomo(t *testing.T) {
	m := newTestManager(t)
	args := m.buildCoreArgs(CoreTypeMihomo, "/path/to/config.json")
	if args[0] != "-f" || args[1] != "/path/to/config.json" {
		t.Fatalf("unexpected mihomo file args: %v", args)
	}
}

// ========== WithStdin StartOption 测试 ==========

func TestWithStdin_SetsConfigData(t *testing.T) {
	sc := &startConfig{}
	data := []byte(`{"inbounds":[],"outbounds":[]}`)
	WithStdin(data)(sc)

	if len(sc.configData) == 0 {
		t.Fatal("expected configData to be set")
	}
	if string(sc.configData) != string(data) {
		t.Fatalf("configData mismatch: got %q", string(sc.configData))
	}
}

func TestStartConfig_DefaultEmpty(t *testing.T) {
	sc := &startConfig{}
	if sc.configData != nil {
		t.Fatal("expected configData to be nil by default")
	}
}

func TestStartCore_AlreadyRunning_ReturnsError(t *testing.T) {
	m := newTestManager(t)
	// 模拟一个正在运行的内核实例
	m.cores[CoreTypeXray] = &coreInstance{
		info: CoreInfo{
			Type:   CoreTypeXray,
			Status: StatusRunning,
			PID:    12345,
		},
	}

	err := m.StartCore(CoreTypeXray, "")
	if err == nil {
		t.Fatal("expected error when core already running")
	}
}

func TestStartCore_BinaryNotFound_ReturnsError(t *testing.T) {
	m := newTestManager(t)
	// 二进制文件不存在
	err := m.StartCore(CoreTypeXray, "")
	if err == nil {
		t.Fatal("expected error when binary not found")
	}
}
