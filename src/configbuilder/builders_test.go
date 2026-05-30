package configbuilder

import (
	"encoding/json"
	"os"
	"testing"

	"v2rayn-go/database"
)

// ========== BuildBytes 测试 ==========

func testProfile() *database.Profile {
	return &database.Profile{
		ProxyProtocol:   "vmess",
		ProxyAddress:    "1.2.3.4",
		ProxyPort:       443,
		ProxyCredential: "00000000-0000-0000-0000-000000000000",
		ProxySecurity:   "auto",
		ProxyNetwork:    "tcp",
		ProxyTLS:        "tls",
		ProxySNI:        "example.com",
	}
}

func TestXrayBuildBytes_ReturnsValidJSON(t *testing.T) {
	builder, ok := GetBuilder("xray")
	if !ok {
		t.Fatal("xray builder not registered")
	}

	data, err := builder.BuildBytes(&BuildConfigParams{
		Profile:   testProfile(),
		Rules:     nil,
		ConfigDir: t.TempDir(),
		SocksPort: 10808,
		HTTPPort:  10809,
	})
	if err != nil {
		t.Fatalf("BuildBytes failed: %v", err)
	}

	// 验证是合法 JSON
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("BuildBytes returned invalid JSON: %v", err)
	}

	// 验证关键字段存在
	inbounds, ok := result["inbounds"].([]any)
	if !ok || len(inbounds) == 0 {
		t.Fatal("inbounds missing or empty")
	}

	outbounds, ok := result["outbounds"].([]any)
	if !ok || len(outbounds) == 0 {
		t.Fatal("outbounds missing or empty")
	}
}

func TestSingboxBuildBytes_ReturnsValidJSON(t *testing.T) {
	builder, ok := GetBuilder("sing-box")
	if !ok {
		t.Fatal("sing-box builder not registered")
	}

	data, err := builder.BuildBytes(&BuildConfigParams{
		Profile:   testProfile(),
		Rules:     nil,
		ConfigDir: t.TempDir(),
		SocksPort: 10808,
		HTTPPort:  10809,
	})
	if err != nil {
		t.Fatalf("BuildBytes failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("BuildBytes returned invalid JSON: %v", err)
	}

	inbounds, ok := result["inbounds"].([]any)
	if !ok || len(inbounds) == 0 {
		t.Fatal("inbounds missing or empty")
	}

	outbounds, ok := result["outbounds"].([]any)
	if !ok || len(outbounds) == 0 {
		t.Fatal("outbounds missing or empty")
	}
}

func TestXrayBuildBytes_MatchesBuildOutput(t *testing.T) {
	builder, ok := GetBuilder("xray")
	if !ok {
		t.Fatal("xray builder not registered")
	}

	params := &BuildConfigParams{
		Profile:   testProfile(),
		Rules:     nil,
		ConfigDir: t.TempDir(),
		SocksPort: 10808,
		HTTPPort:  10809,
	}

	// BuildBytes 结果
	bytesData, err := builder.BuildBytes(params)
	if err != nil {
		t.Fatalf("BuildBytes failed: %v", err)
	}

	// Build 结果
	configPath, err := builder.Build(params)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	fileData, err := readFileContent(configPath)
	if err != nil {
		t.Fatalf("failed to read build output: %v", err)
	}

	// 两者应该是相同的 JSON 结构
	var a, b map[string]any
	json.Unmarshal(bytesData, &a)
	json.Unmarshal(fileData, &b)

	aInbounds := a["inbounds"].([]any)
	bInbounds := b["inbounds"].([]any)
	if len(aInbounds) != len(bInbounds) {
		t.Fatalf("inbounds count mismatch: BuildBytes=%d, Build=%d", len(aInbounds), len(bInbounds))
	}
}

func TestBuildBytes_InvalidProtocol_ReturnsError(t *testing.T) {
	builder, ok := GetBuilder("xray")
	if !ok {
		t.Fatal("xray builder not registered")
	}

	_, err := builder.BuildBytes(&BuildConfigParams{
		Profile: &database.Profile{
			ProxyProtocol: "unsupported_proto",
			ProxyAddress:  "1.2.3.4",
			ProxyPort:     443,
		},
		Rules:     nil,
		ConfigDir: t.TempDir(),
		SocksPort: 10808,
		HTTPPort:  10809,
	})
	if err == nil {
		t.Fatal("expected error for unsupported protocol, got nil")
	}
}

func TestBuildBytes_NilProfile_ReturnsError(t *testing.T) {
	builder, ok := GetBuilder("xray")
	if !ok {
		t.Fatal("xray builder not registered")
	}

	_, err := builder.BuildBytes(&BuildConfigParams{
		Profile:   nil,
		Rules:     nil,
		ConfigDir: t.TempDir(),
		SocksPort: 10808,
		HTTPPort:  10809,
	})
	if err == nil {
		t.Fatal("expected error for nil profile, got nil")
	}
}

// readFileContent 读取文件内容（测试辅助函数）
func readFileContent(path string) ([]byte, error) {
	return os.ReadFile(path)
}
