package parser

import (
	"testing"
)

func TestParseWireGuard_Basic(t *testing.T) {
	link := "wireguard://privateKey123@host.com:51820?public_key=pubkey123&reserved=1,2,3&address=10.0.0.2/32#TestWG"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxyProtocol != "wireguard" {
		t.Fatalf("expected protocol 'wireguard', got '%s'", profile.ProxyProtocol)
	}
	if profile.Name != "TestWG" {
		t.Fatalf("expected name 'TestWG', got '%s'", profile.Name)
	}
	if profile.ProxyAddress != "host.com" {
		t.Fatalf("expected address 'host.com', got '%s'", profile.ProxyAddress)
	}
	if profile.ProxyPort != 51820 {
		t.Fatalf("expected port 51820, got %d", profile.ProxyPort)
	}
	if profile.ProxyCredential != "privateKey123" {
		t.Fatalf("expected privateKey 'privateKey123', got '%s'", profile.ProxyCredential)
	}
	if profile.ProxyPublicKey != "pubkey123" {
		t.Fatalf("expected publicKey 'pubkey123', got '%s'", profile.ProxyPublicKey)
	}
	if profile.ProxyPath != "1,2,3" {
		t.Fatalf("expected reserved '1,2,3', got '%s'", profile.ProxyPath)
	}
	if profile.ProxyHost != "10.0.0.2/32" {
		t.Fatalf("expected address '10.0.0.2/32', got '%s'", profile.ProxyHost)
	}
}

func TestParseWireGuard_WgScheme(t *testing.T) {
	link := "wg://key@1.2.3.4:51820?pk=shortpubkey#WGShort"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxyProtocol != "wireguard" {
		t.Fatalf("expected protocol 'wireguard', got '%s'", profile.ProxyProtocol)
	}
	if profile.ProxyPublicKey != "shortpubkey" {
		t.Fatalf("expected publicKey 'shortpubkey', got '%s'", profile.ProxyPublicKey)
	}
	if profile.Name != "WGShort" {
		t.Fatalf("expected name 'WGShort', got '%s'", profile.Name)
	}
}

func TestParseWireGuard_ShortParamNames(t *testing.T) {
	link := "wireguard://key@host.com:51820?pk=mypubkey&reserved=0,0,0&addr=10.0.0.3/32#Test"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxyPublicKey != "mypubkey" {
		t.Fatalf("expected publicKey 'mypubkey', got '%s'", profile.ProxyPublicKey)
	}
	if profile.ProxyHost != "10.0.0.3/32" {
		t.Fatalf("expected addr '10.0.0.3/32', got '%s'", profile.ProxyHost)
	}
}

func TestParseWireGuard_DefaultName(t *testing.T) {
	link := "wireguard://key@10.0.0.1:51820?pk=pub"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Default name should be wg-address:port
	if profile.Name != "wg-10.0.0.1:51820" {
		t.Fatalf("expected default name 'wg-10.0.0.1:51820', got '%s'", profile.Name)
	}
}

func TestParseWireGuard_DefaultPort(t *testing.T) {
	link := "wireguard://key@host.com?pk=pub#Test"

	profile, err := ParseLink(link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.ProxyPort != 51820 {
		t.Fatalf("expected default port 51820, got %d", profile.ProxyPort)
	}
}
