package parser

import (
	"testing"
)

func TestParseTrojan_Basic(t *testing.T) {
	link := "trojan://password123@example.com:443?sni=example.com&fp=chrome#TestTrojan"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxyProtocol != "trojan" {
		t.Fatalf("expected protocol 'trojan', got '%s'", profile.ProxyProtocol)
	}
	if profile.Name != "TestTrojan" {
		t.Fatalf("expected name 'TestTrojan', got '%s'", profile.Name)
	}
	if profile.ProxyAddress != "example.com" {
		t.Fatalf("expected address 'example.com', got '%s'", profile.ProxyAddress)
	}
	if profile.ProxyPort != 443 {
		t.Fatalf("expected port 443, got %d", profile.ProxyPort)
	}
	if profile.ProxyCredential != "password123" {
		t.Fatalf("expected credential 'password123', got '%s'", profile.ProxyCredential)
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

func TestParseTrojan_WithNetwork(t *testing.T) {
	link := "trojan://pass@host.com:443?type=ws&host=ws.host.com&path=/trojan-ws#WS Trojan"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxyNetwork != "ws" {
		t.Fatalf("expected network 'ws', got '%s'", profile.ProxyNetwork)
	}
	if profile.ProxyHost != "ws.host.com" {
		t.Fatalf("expected host 'ws.host.com', got '%s'", profile.ProxyHost)
	}
	if profile.ProxyPath != "/trojan-ws" {
		t.Fatalf("expected path '/trojan-ws', got '%s'", profile.ProxyPath)
	}
}

func TestParseTrojan_AllowInsecure(t *testing.T) {
	tests := []struct {
		name     string
		param    string
		expected bool
	}{
		{"true string", "true", true},
		{"1 string", "1", true},
		{"0 string", "0", false},
		{"false string", "false", false},
		{"empty", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			link := "trojan://pass@host.com:443?allowInsecure=" + tc.param + "#Test"
			profile, err := ParseLink(link)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if profile.ProxyAllowInsecure != tc.expected {
				t.Fatalf("expected allowInsecure=%v, got %v", tc.expected, profile.ProxyAllowInsecure)
			}
		})
	}
}

func TestParseTrojan_Reality(t *testing.T) {
	link := "trojan://pass@host.com:443?security=reality&pbk=publickey&sid=shortid#Reality"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxyTLS != "reality" {
		t.Fatalf("expected TLS 'reality', got '%s'", profile.ProxyTLS)
	}
	if profile.ProxyPublicKey != "publickey" {
		t.Fatalf("expected publicKey 'publickey', got '%s'", profile.ProxyPublicKey)
	}
	if profile.ProxyShortID != "shortid" {
		t.Fatalf("expected shortID 'shortid', got '%s'", profile.ProxyShortID)
	}
}

func TestParseTrojan_Defaults(t *testing.T) {
	link := "trojan://pass@1.2.3.4:8443"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxyNetwork != "tcp" {
		t.Fatalf("expected default network 'tcp', got '%s'", profile.ProxyNetwork)
	}
	if profile.ProxyTLS != "tls" {
		t.Fatalf("expected default TLS 'tls', got '%s'", profile.ProxyTLS)
	}
	if profile.Name != "1.2.3.4:8443" {
		t.Fatalf("expected default name '1.2.3.4:8443', got '%s'", profile.Name)
	}
}

func TestParseTrojan_WithFlow(t *testing.T) {
	link := "trojan://pass@host.com:443?flow=xtls-rprx-vision#Flow"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxyFlow != "xtls-rprx-vision" {
		t.Fatalf("expected flow 'xtls-rprx-vision', got '%s'", profile.ProxyFlow)
	}
}
