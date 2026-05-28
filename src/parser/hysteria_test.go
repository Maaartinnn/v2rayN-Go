package parser

import (
	"testing"
)

// ==================== Hysteria2 ====================

func TestParseHysteria2_Basic(t *testing.T) {
	link := "hysteria2://password@example.com:443?sni=example.com#TestHy2"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxyProtocol != "hysteria2" {
		t.Fatalf("expected protocol 'hysteria2', got '%s'", profile.ProxyProtocol)
	}
	if profile.Name != "TestHy2" {
		t.Fatalf("expected name 'TestHy2', got '%s'", profile.Name)
	}
	if profile.ProxyAddress != "example.com" {
		t.Fatalf("expected address 'example.com', got '%s'", profile.ProxyAddress)
	}
	if profile.ProxyPort != 443 {
		t.Fatalf("expected port 443, got %d", profile.ProxyPort)
	}
	if profile.ProxyCredential != "password" {
		t.Fatalf("expected credential 'password', got '%s'", profile.ProxyCredential)
	}
	if profile.ProxyTLS != "tls" {
		t.Fatalf("expected TLS 'tls', got '%s'", profile.ProxyTLS)
	}
	if profile.ProxySNI != "example.com" {
		t.Fatalf("expected SNI 'example.com', got '%s'", profile.ProxySNI)
	}
}

func TestParseHysteria2_Hy2Scheme(t *testing.T) {
	// hy2:// 短格式
	link := "hy2://pass@1.2.3.4:8443#Hy2Short"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxyProtocol != "hysteria2" {
		t.Fatalf("expected protocol 'hysteria2', got '%s'", profile.ProxyProtocol)
	}
	if profile.Name != "Hy2Short" {
		t.Fatalf("expected name 'Hy2Short', got '%s'", profile.Name)
	}
}

func TestParseHysteria2_Insecure(t *testing.T) {
	tests := []struct {
		name     string
		param    string
		expected bool
	}{
		{"insecure=1", "1", true},
		{"insecure=true", "true", true},
		{"insecure=0", "0", false},
		{"insecure=empty", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			link := "hysteria2://pass@host.com:443?insecure=" + tc.param + "#Test"
			profile, err := ParseLink(link)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if profile.ProxyAllowInsecure != tc.expected {
				t.Fatalf("expected insecure=%v, got %v", tc.expected, profile.ProxyAllowInsecure)
			}
		})
	}
}

func TestParseHysteria2_WithObfs(t *testing.T) {
	link := "hysteria2://pass@host.com:443?obfs=salamander&obfs-password=obfspass#Obfs"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxyNetwork != "salamander" {
		t.Fatalf("expected obfs 'salamander', got '%s'", profile.ProxyNetwork)
	}
	if profile.ProxyPath != "obfspass" {
		t.Fatalf("expected obfs-password 'obfspass', got '%s'", profile.ProxyPath)
	}
}

func TestParseHysteria2_DefaultSNIsHost(t *testing.T) {
	link := "hysteria2://pass@myhost.com:443#Test"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxySNI != "myhost.com" {
		t.Fatalf("expected default SNI 'myhost.com', got '%s'", profile.ProxySNI)
	}
}

// ==================== Hysteria ====================

func TestParseHysteria_Basic(t *testing.T) {
	link := "hysteria://host.com:443?auth=myauth&sni=host.com&insecure=1#TestHysteria"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxyProtocol != "hysteria" {
		t.Fatalf("expected protocol 'hysteria', got '%s'", profile.ProxyProtocol)
	}
	if profile.Name != "TestHysteria" {
		t.Fatalf("expected name 'TestHysteria', got '%s'", profile.Name)
	}
	if profile.ProxyCredential != "myauth" {
		t.Fatalf("expected credential 'myauth', got '%s'", profile.ProxyCredential)
	}
	if profile.ProxyAllowInsecure != true {
		t.Fatal("expected insecure=true")
	}
}

func TestParseHysteria_AuthFromUserinfo(t *testing.T) {
	link := "hysteria://myauth@host.com:443#Test"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxyCredential != "myauth" {
		t.Fatalf("expected credential 'myauth', got '%s'", profile.ProxyCredential)
	}
}

func TestParseHysteria_WithObfs(t *testing.T) {
	link := "hysteria://host.com:443?obfs=xplus&obfs-password=obfspw#Obfs"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxyNetwork != "xplus" {
		t.Fatalf("expected obfs 'xplus', got '%s'", profile.ProxyNetwork)
	}
	if profile.ProxyPath != "obfspw" {
		t.Fatalf("expected obfs-password 'obfspw', got '%s'", profile.ProxyPath)
	}
}

// ==================== TUIC ====================

func TestParseTuic_Basic(t *testing.T) {
	link := "tuic://uuid:password@host.com:443?sni=host.com&congestion_control=bbr#TestTUIC"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxyProtocol != "tuic" {
		t.Fatalf("expected protocol 'tuic', got '%s'", profile.ProxyProtocol)
	}
	if profile.Name != "TestTUIC" {
		t.Fatalf("expected name 'TestTUIC', got '%s'", profile.Name)
	}
	if profile.ProxyCredential != "uuid" {
		t.Fatalf("expected credential 'uuid', got '%s'", profile.ProxyCredential)
	}
	if profile.ProxySecurity != "password" {
		t.Fatalf("expected security(password) 'password', got '%s'", profile.ProxySecurity)
	}
	if profile.ProxyTLS != "tls" {
		t.Fatalf("expected TLS 'tls', got '%s'", profile.ProxyTLS)
	}
	if profile.ProxyNetwork != "bbr" {
		t.Fatalf("expected congestion_control 'bbr', got '%s'", profile.ProxyNetwork)
	}
}

func TestParseTuic_AllowInsecure(t *testing.T) {
	link := "tuic://uuid:pass@host.com:443?allow_insecure=1#Test"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !profile.ProxyAllowInsecure {
		t.Fatal("expected allow_insecure=true")
	}
}
