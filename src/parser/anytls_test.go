package parser

import (
	"testing"
)

func TestParseAnytls_Basic(t *testing.T) {
	link := "anytls://mypassword@example.com:443?sni=example.com&fp=firefox#TestAnyTLS"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxyProtocol != "anytls" {
		t.Fatalf("expected protocol 'anytls', got '%s'", profile.ProxyProtocol)
	}
	if profile.Name != "TestAnyTLS" {
		t.Fatalf("expected name 'TestAnyTLS', got '%s'", profile.Name)
	}
	if profile.ProxyAddress != "example.com" {
		t.Fatalf("expected address 'example.com', got '%s'", profile.ProxyAddress)
	}
	if profile.ProxyPort != 443 {
		t.Fatalf("expected port 443, got %d", profile.ProxyPort)
	}
	if profile.ProxyCredential != "mypassword" {
		t.Fatalf("expected credential 'mypassword', got '%s'", profile.ProxyCredential)
	}
	if profile.ProxyTLS != "tls" {
		t.Fatalf("expected TLS 'tls', got '%s'", profile.ProxyTLS)
	}
	if profile.ProxySNI != "example.com" {
		t.Fatalf("expected SNI 'example.com', got '%s'", profile.ProxySNI)
	}
	if profile.ProxyFingerprint != "firefox" {
		t.Fatalf("expected fingerprint 'firefox', got '%s'", profile.ProxyFingerprint)
	}
}

func TestParseAnytls_AllowInsecure(t *testing.T) {
	tests := []struct {
		name     string
		param    string
		expected bool
	}{
		{"allowInsecure=1", "1", true},
		{"allowInsecure=true", "true", true},
		{"allowInsecure=0", "0", false},
		{"empty", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			link := "anytls://pass@host.com:443?allowInsecure=" + tc.param + "#Test"
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

func TestParseAnytls_DefaultSNIsHost(t *testing.T) {
	link := "anytls://pass@myserver.com:443#Test"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxySNI != "myserver.com" {
		t.Fatalf("expected default SNI 'myserver.com', got '%s'", profile.ProxySNI)
	}
}

func TestParseAnytls_DefaultName(t *testing.T) {
	link := "anytls://pass@1.2.3.4:8443"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.Name != "anytls-1.2.3.4:8443" {
		t.Fatalf("expected default name 'anytls-1.2.3.4:8443', got '%s'", profile.Name)
	}
}

func TestParseAnytls_DefaultPort(t *testing.T) {
	link := "anytls://pass@host.com#Test"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxyPort != 443 {
		t.Fatalf("expected default port 443, got %d", profile.ProxyPort)
	}
}
