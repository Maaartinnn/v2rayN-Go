package parser

import (
	"testing"
)

func TestParseVless_Basic(t *testing.T) {
	link := "vless://a3482e88-686a-4a58-8126-99c95f6f5091@example.com:443?type=tcp&security=tls&sni=example.com&fp=chrome#TestVless"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxyProtocol != "vless" {
		t.Fatalf("expected protocol 'vless', got '%s'", profile.ProxyProtocol)
	}
	if profile.Name != "TestVless" {
		t.Fatalf("expected name 'TestVless', got '%s'", profile.Name)
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

func TestParseVless_WithFlow(t *testing.T) {
	link := "vless://uuid@1.2.3.4:1234?type=tcp&security=reality&flow=xtls-rprx-vision&sni=www.google.com&pbk=publickey&sid=shortid#Reality"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxyFlow != "xtls-rprx-vision" {
		t.Fatalf("expected flow 'xtls-rprx-vision', got '%s'", profile.ProxyFlow)
	}
	if profile.ProxyPublicKey != "publickey" {
		t.Fatalf("expected publicKey 'publickey', got '%s'", profile.ProxyPublicKey)
	}
	if profile.ProxyShortID != "shortid" {
		t.Fatalf("expected shortID 'shortid', got '%s'", profile.ProxyShortID)
	}
}

func TestParseVless_WS(t *testing.T) {
	link := "vless://uuid@host.com:80?type=ws&security=none&host=ws.host.com&path=/websocket#WSNode"

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
	if profile.ProxyPath != "/websocket" {
		t.Fatalf("expected path '/websocket', got '%s'", profile.ProxyPath)
	}
}

func TestParseVless_Defaults(t *testing.T) {
	link := "vless://uuid@1.2.3.4:8080"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxyNetwork != "tcp" {
		t.Fatalf("expected default network 'tcp', got '%s'", profile.ProxyNetwork)
	}
	if profile.Name != "1.2.3.4:8080" {
		t.Fatalf("expected default name '1.2.3.4:8080', got '%s'", profile.Name)
	}
}

func TestParseVless_WithSeed(t *testing.T) {
	link := "vless://uuid@example.com:443?type=quic&security=tls&seed=myseed#QUIC"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxySeed != "myseed" {
		t.Fatalf("expected seed 'myseed', got '%s'", profile.ProxySeed)
	}
}
