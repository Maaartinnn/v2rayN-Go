package parser

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

// 构造 VMess JSON 格式的 base64 链接
func makeVmessLink(vj vmessJSON) string {
	data, _ := json.Marshal(vj)
	return "vmess://" + base64.StdEncoding.EncodeToString(data)
}

func TestParseVmess_JSON_Standard(t *testing.T) {
	link := makeVmessLink(vmessJSON{
		V:    "2",
		Ps:   "TestVMess",
		Add:  "example.com",
		Port: 443,
		ID:   "a3482e88-686a-4a58-8126-99c95f6f5091",
		Aid:  0,
		Scy:  "auto",
		Net:  "tcp",
		TLS:  "tls",
		SNI:  "example.com",
		Fp:   "chrome",
	})

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxyProtocol != "vmess" {
		t.Fatalf("expected protocol 'vmess', got '%s'", profile.ProxyProtocol)
	}
	if profile.Name != "TestVMess" {
		t.Fatalf("expected name 'TestVMess', got '%s'", profile.Name)
	}
	if profile.ProxyAddress != "example.com" {
		t.Fatalf("expected address 'example.com', got '%s'", profile.ProxyAddress)
	}
	if profile.ProxyPort != 443 {
		t.Fatalf("expected port 443, got %d", profile.ProxyPort)
	}
	if profile.ProxyCredential != "a3482e88-686a-4a58-8126-99c95f6f5091" {
		t.Fatalf("expected UUID credential, got '%s'", profile.ProxyCredential)
	}
	if profile.ProxyAlterID != 0 {
		t.Fatalf("expected alterID 0, got %d", profile.ProxyAlterID)
	}
	if profile.ProxySecurity != "auto" {
		t.Fatalf("expected security 'auto', got '%s'", profile.ProxySecurity)
	}
	if profile.ProxyNetwork != "tcp" {
		t.Fatalf("expected network 'tcp', got '%s'", profile.ProxyNetwork)
	}
	if profile.ProxyTLS != "tls" {
		t.Fatalf("expected TLS 'tls', got '%s'", profile.ProxyTLS)
	}
	if profile.ProxySNI != "example.com" {
		t.Fatalf("expected SNI 'example.com', got '%s'", profile.ProxySNI)
	}
	if profile.ProxyFingerprint != "chrome" {
		t.Fatalf("expected fingerprint 'chrome', got '%s'", profile.ProxyFingerprint)
	}
}

func TestParseVmess_JSON_StringPort(t *testing.T) {
	link := makeVmessLink(vmessJSON{
		V:    "2",
		Ps:   "StringPort",
		Add:  "1.2.3.4",
		Port: "8080",
		ID:   "uuid-test",
		Aid:  "2",
		Net:  "ws",
		Host: "ws.example.com",
		Path: "/ws",
	})

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxyPort != 8080 {
		t.Fatalf("expected port 8080, got %d", profile.ProxyPort)
	}
	if profile.ProxyAlterID != 2 {
		t.Fatalf("expected alterID 2, got %d", profile.ProxyAlterID)
	}
	if profile.ProxyNetwork != "ws" {
		t.Fatalf("expected network 'ws', got '%s'", profile.ProxyNetwork)
	}
	if profile.ProxyHost != "ws.example.com" {
		t.Fatalf("expected host 'ws.example.com', got '%s'", profile.ProxyHost)
	}
	if profile.ProxyPath != "/ws" {
		t.Fatalf("expected path '/ws', got '%s'", profile.ProxyPath)
	}
}

func TestParseVmess_JSON_Defaults(t *testing.T) {
	// 测试默认值：security 默认 auto，network 默认 tcp，name 默认 address:port
	link := makeVmessLink(vmessJSON{
		Add:  "10.0.0.1",
		Port: 1234,
		ID:   "test-uuid",
	})

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxySecurity != "auto" {
		t.Fatalf("expected default security 'auto', got '%s'", profile.ProxySecurity)
	}
	if profile.ProxyNetwork != "tcp" {
		t.Fatalf("expected default network 'tcp', got '%s'", profile.ProxyNetwork)
	}
	if profile.Name != "10.0.0.1:1234" {
		t.Fatalf("expected default name '10.0.0.1:1234', got '%s'", profile.Name)
	}
}

func TestParseVmess_InvalidBase64(t *testing.T) {
	_, err := ParseLink("vmess://!!!invalid-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestParseVmess_InvalidJSON(t *testing.T) {
	// Valid base64 but not valid JSON
	invalidJSON := base64.StdEncoding.EncodeToString([]byte("not json"))
	_, err := ParseLink("vmess://" + invalidJSON)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
