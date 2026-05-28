package parser

import (
	"encoding/base64"
	"testing"
)

func TestParseShadowsocks_URI(t *testing.T) {
	// ss://base64(method:password)@host:port#name
	userInfo := base64.StdEncoding.EncodeToString([]byte("aes-256-gcm:password123"))
	link := "ss://" + userInfo + "@example.com:8388#TestSS"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxyProtocol != "shadowsocks" {
		t.Fatalf("expected protocol 'shadowsocks', got '%s'", profile.ProxyProtocol)
	}
	if profile.Name != "TestSS" {
		t.Fatalf("expected name 'TestSS', got '%s'", profile.Name)
	}
	if profile.ProxyAddress != "example.com" {
		t.Fatalf("expected address 'example.com', got '%s'", profile.ProxyAddress)
	}
	if profile.ProxyPort != 8388 {
		t.Fatalf("expected port 8388, got %d", profile.ProxyPort)
	}
	if profile.ProxySecurity != "aes-256-gcm" {
		t.Fatalf("expected method 'aes-256-gcm', got '%s'", profile.ProxySecurity)
	}
	if profile.ProxyCredential != "password123" {
		t.Fatalf("expected password 'password123', got '%s'", profile.ProxyCredential)
	}
}

func TestParseShadowsocks_FullBase64(t *testing.T) {
	// ss://base64(method:password@host:port)#name
	plain := "chacha20-ietf-poly1305:mypass@ss.example.com:12345"
	encoded := base64.StdEncoding.EncodeToString([]byte(plain))
	link := "ss://" + encoded + "#FullSS"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxyProtocol != "shadowsocks" {
		t.Fatalf("expected protocol 'shadowsocks', got '%s'", profile.ProxyProtocol)
	}
	if profile.ProxySecurity != "chacha20-ietf-poly1305" {
		t.Fatalf("expected method 'chacha20-ietf-poly1305', got '%s'", profile.ProxySecurity)
	}
	if profile.ProxyCredential != "mypass" {
		t.Fatalf("expected password 'mypass', got '%s'", profile.ProxyCredential)
	}
	if profile.ProxyAddress != "ss.example.com" {
		t.Fatalf("expected address 'ss.example.com', got '%s'", profile.ProxyAddress)
	}
	if profile.ProxyPort != 12345 {
		t.Fatalf("expected port 12345, got %d", profile.ProxyPort)
	}
	if profile.Name != "FullSS" {
		t.Fatalf("expected name 'FullSS', got '%s'", profile.Name)
	}
}

func TestParseShadowsocks_WithPlugin(t *testing.T) {
	userInfo := base64.StdEncoding.EncodeToString([]byte("aes-128-gcm:pass"))
	link := "ss://" + userInfo + "@host.com:443?plugin=obfs-local%3Bobfs%3Dhttp#PluginSS"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxyPath == "" {
		t.Fatal("expected plugin in ProxyPath, got empty")
	}
}

func TestParseShadowsocks_DefaultName(t *testing.T) {
	userInfo := base64.StdEncoding.EncodeToString([]byte("rc4-md5:pass"))
	link := "ss://" + userInfo + "@10.0.0.1:9999"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.Name != "10.0.0.1:9999" {
		t.Fatalf("expected default name '10.0.0.1:9999', got '%s'", profile.Name)
	}
}

func TestParseShadowsocks_InvalidBase64(t *testing.T) {
	_, err := ParseLink("ss://!!!invalid!!!@host.com:443")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

// ==================== ShadowsocksR ====================

func TestParseShadowsocksR_Basic(t *testing.T) {
	// ssr://base64(host:port:protocol:method:obfs:base64pass/?params)
	pass := base64.StdEncoding.EncodeToString([]byte("mypassword"))
	remarks := base64.StdEncoding.EncodeToString([]byte("SSRNode"))
	plain := "ss.example.com:12345:origin:aes-256-cfb:plain:" + pass + "/?remarks=" + remarks
	encoded := base64.RawStdEncoding.EncodeToString([]byte(plain))
	link := "ssr://" + encoded

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxyProtocol != "shadowsocksr" {
		t.Fatalf("expected protocol 'shadowsocksr', got '%s'", profile.ProxyProtocol)
	}
	if profile.ProxyAddress != "ss.example.com" {
		t.Fatalf("expected address 'ss.example.com', got '%s'", profile.ProxyAddress)
	}
	if profile.ProxyPort != 12345 {
		t.Fatalf("expected port 12345, got %d", profile.ProxyPort)
	}
	if profile.ProxySecurity != "aes-256-cfb" {
		t.Fatalf("expected method 'aes-256-cfb', got '%s'", profile.ProxySecurity)
	}
	if profile.ProxyCredential != "mypassword" {
		t.Fatalf("expected password 'mypassword', got '%s'", profile.ProxyCredential)
	}
	if profile.ProxyNetwork != "origin" {
		t.Fatalf("expected protocol 'origin', got '%s'", profile.ProxyNetwork)
	}
	if profile.ProxyPath != "plain" {
		t.Fatalf("expected obfs 'plain', got '%s'", profile.ProxyPath)
	}
	if profile.Name != "SSRNode" {
		t.Fatalf("expected name 'SSRNode', got '%s'", profile.Name)
	}
}

func TestParseShadowsocksR_InvalidFormat(t *testing.T) {
	// Too few parts
	plain := "host:port:proto"
	encoded := base64.StdEncoding.EncodeToString([]byte(plain))
	_, err := ParseLink("ssr://" + encoded)
	if err == nil {
		t.Fatal("expected error for invalid SSR format")
	}
}
