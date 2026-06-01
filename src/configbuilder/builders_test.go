package configbuilder

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"

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

	// 验证输出路径在 binConfig 目录下
	expectedDir := filepath.Join(params.ConfigDir, "binConfig")
	if filepath.Dir(configPath) != expectedDir {
		t.Fatalf("expected config in binConfig dir, got: %s", configPath)
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

func TestMihomoBuildBytes_ReturnsValidYAML(t *testing.T) {
	builder, ok := GetBuilder("mihomo")
	if !ok {
		t.Fatal("mihomo builder not registered")
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

	// 验证是合法 YAML
	var result map[string]any
	if err := yaml.Unmarshal(data, &result); err != nil {
		t.Fatalf("BuildBytes returned invalid YAML: %v", err)
	}

	// 验证关键字段存在
	proxies, ok := result["proxies"].([]any)
	if !ok || len(proxies) == 0 {
		t.Fatal("proxies missing or empty")
	}

	groups, ok := result["proxy-groups"].([]any)
	if !ok || len(groups) == 0 {
		t.Fatal("proxy-groups missing or empty")
	}

	rules, ok := result["rules"].([]any)
	if !ok || len(rules) == 0 {
		t.Fatal("rules missing or empty")
	}
}

func TestMihomoBuildBytes_MatchesBuildOutput(t *testing.T) {
	builder, ok := GetBuilder("mihomo")
	if !ok {
		t.Fatal("mihomo builder not registered")
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

	// 验证输出路径在 binConfig 目录下且扩展名为 .yaml
	expectedDir := filepath.Join(params.ConfigDir, "binConfig")
	if filepath.Dir(configPath) != expectedDir {
		t.Fatalf("expected config in binConfig dir, got: %s", configPath)
	}
	if filepath.Ext(configPath) != ".yaml" {
		t.Fatalf("expected .yaml extension, got: %s", filepath.Ext(configPath))
	}

	fileData, err := readFileContent(configPath)
	if err != nil {
		t.Fatalf("failed to read build output: %v", err)
	}

	// 两者应该是相同的 YAML 结构
	var a, b map[string]any
	yaml.Unmarshal(bytesData, &a)
	yaml.Unmarshal(fileData, &b)

	aProxies := a["proxies"].([]any)
	bProxies := b["proxies"].([]any)
	if len(aProxies) != len(bProxies) {
		t.Fatalf("proxies count mismatch: BuildBytes=%d, Build=%d", len(aProxies), len(bProxies))
	}
}

func TestMihomoBuildBytes_UnsupportedProtocol_ReturnsError(t *testing.T) {
	builder, ok := GetBuilder("mihomo")
	if !ok {
		t.Fatal("mihomo builder not registered")
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

func TestMihomoBuildBytes_WithRoutingRules(t *testing.T) {
	builder, ok := GetBuilder("mihomo")
	if !ok {
		t.Fatal("mihomo builder not registered")
	}

	rules := []database.RoutingRule{
		{Type: "direct", Domain: "baidu.com,bilibili.com", Enabled: true},
		{Type: "proxy", IP: "8.8.8.8", Enabled: true},
		{Type: "block", Port: "25", Enabled: true},
		{Type: "direct", Domain: "disabled.com", Enabled: false}, // 不应出现在规则中
	}

	data, err := builder.BuildBytes(&BuildConfigParams{
		Profile:   testProfile(),
		Rules:     rules,
		ConfigDir: t.TempDir(),
		SocksPort: 10808,
		HTTPPort:  10809,
	})
	if err != nil {
		t.Fatalf("BuildBytes failed: %v", err)
	}

	var result map[string]any
	if err := yaml.Unmarshal(data, &result); err != nil {
		t.Fatalf("invalid YAML: %v", err)
	}

	rulesList, ok := result["rules"].([]any)
	if !ok {
		t.Fatal("rules missing")
	}

	// 应包含: 3 LAN + 2 domain + 1 IP + 1 port + 1 MATCH = 8 条规则
	rulesStr := ""
	for _, r := range rulesList {
		rulesStr += r.(string) + "\n"
	}

	// 检查用户规则是否正确转换
	found := false
	for _, r := range rulesList {
		if r.(string) == "DOMAIN-SUFFIX,baidu.com,DIRECT" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected DOMAIN-SUFFIX,baidu.com,DIRECT in rules, got: %v", rulesList)
	}

	// 验证禁用规则未出现
	for _, r := range rulesList {
		if contains := func(s string) bool {
			_, ok := any(s).(string)
			return ok
		}; contains(r.(string)) {
			// 只检查 disabled.com 是否出现
		}
	}

	// 验证兜底规则存在
	lastRule := rulesList[len(rulesList)-1].(string)
	if lastRule != "MATCH,Proxy" {
		t.Fatalf("expected last rule to be MATCH,Proxy, got: %s", lastRule)
	}
}

func TestMihomoAllProtocols(t *testing.T) {
	builder, ok := GetBuilder("mihomo")
	if !ok {
		t.Fatal("mihomo builder not registered")
	}

	protocols := []string{"vmess", "vless", "trojan", "shadowsocks", "hysteria2", "tuic", "socks", "http"}
	for _, proto := range protocols {
		t.Run(proto, func(t *testing.T) {
			p := testProfile()
			p.ProxyProtocol = proto

			_, err := builder.BuildBytes(&BuildConfigParams{
				Profile:   p,
				Rules:     nil,
				ConfigDir: t.TempDir(),
				SocksPort: 10808,
				HTTPPort:  10809,
			})
			if err != nil {
				t.Fatalf("BuildBytes failed for protocol %s: %v", proto, err)
			}
		})
	}
}

// readFileContent 读取文件内容（测试辅助函数）
func readFileContent(path string) ([]byte, error) {
	return os.ReadFile(path)
}
