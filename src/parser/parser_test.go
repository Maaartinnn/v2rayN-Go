package parser

import (
	"testing"
)

func TestParseLink_Empty(t *testing.T) {
	_, err := ParseLink("")
	if err == nil {
		t.Fatal("expected error for empty link")
	}
}

func TestParseLink_Whitespace(t *testing.T) {
	_, err := ParseLink("   \n\t  ")
	if err == nil {
		t.Fatal("expected error for whitespace-only link")
	}
}

func TestParseLink_UnsupportedProtocol(t *testing.T) {
	_, err := ParseLink("ftp://example.com")
	if err == nil {
		t.Fatal("expected error for unsupported protocol")
	}
}

func TestParseLinks_MultipleValid(t *testing.T) {
	links := []string{
		"vless://uuid@example.com:443?type=tcp&security=tls#TestVless",
		"trojan://password@example.com:443#TestTrojan",
	}
	profiles, err := ParseLinks(links)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}
}

func TestParseLinks_SkipEmpty(t *testing.T) {
	links := []string{
		"",
		"   ",
		"vless://uuid@example.com:443?type=tcp&security=tls#TestVless",
	}
	profiles, err := ParseLinks(links)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}
}

func TestParseLinks_AllFail(t *testing.T) {
	links := []string{
		"ftp://invalid1",
		"ftp://invalid2",
	}
	_, err := ParseLinks(links)
	if err == nil {
		t.Fatal("expected error when all links fail")
	}
}

func TestTruncate_Short(t *testing.T) {
	result := truncate("hello", 10)
	if result != "hello" {
		t.Fatalf("expected 'hello', got '%s'", result)
	}
}

func TestTruncate_Long(t *testing.T) {
	result := truncate("hello world this is long", 10)
	if result != "hello worl..." {
		t.Fatalf("expected 'hello worl...', got '%s'", result)
	}
}

func TestTruncate_Exact(t *testing.T) {
	result := truncate("12345", 5)
	if result != "12345" {
		t.Fatalf("expected '12345', got '%s'", result)
	}
}
