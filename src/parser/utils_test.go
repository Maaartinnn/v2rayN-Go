package parser

import (
	"net/url"
	"testing"
)

// ==================== base64Decode ====================

func TestBase64Decode_StdEncoding(t *testing.T) {
	// "hello world" in standard base64
	input := "aGVsbG8gd29ybGQ="
	result, err := base64Decode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello world" {
		t.Fatalf("expected 'hello world', got '%s'", result)
	}
}

func TestBase64Decode_URLEncoding(t *testing.T) {
	// "hello?world" in URL-safe base64
	input := "aGVsbG8_d29ybGQ="
	result, err := base64Decode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello?world" {
		t.Fatalf("expected 'hello?world', got '%s'", result)
	}
}

func TestBase64Decode_RawStdEncoding(t *testing.T) {
	// "test" in raw standard base64 (no padding)
	input := "dGVzdA"
	result, err := base64Decode(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "test" {
		t.Fatalf("expected 'test', got '%s'", result)
	}
}

func TestBase64Decode_Invalid(t *testing.T) {
	_, err := base64Decode("!!!not-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

// ==================== parseIntSafe ====================

func TestParseIntSafe_Valid(t *testing.T) {
	result := parseIntSafe("443", 80)
	if result != 443 {
		t.Fatalf("expected 443, got %d", result)
	}
}

func TestParseIntSafe_Empty(t *testing.T) {
	result := parseIntSafe("", 80)
	if result != 80 {
		t.Fatalf("expected 80, got %d", result)
	}
}

func TestParseIntSafe_Invalid(t *testing.T) {
	result := parseIntSafe("abc", 80)
	if result != 80 {
		t.Fatalf("expected 80, got %d", result)
	}
}

func TestParseIntSafe_Zero(t *testing.T) {
	result := parseIntSafe("0", 80)
	if result != 0 {
		t.Fatalf("expected 0, got %d", result)
	}
}

// ==================== extractNameFromFragment ====================

func TestExtractNameFromFragment_WithFragment(t *testing.T) {
	u, _ := url.Parse("vmess://abc@host:443#MyServer")
	result := extractNameFromFragment(u)
	if result != "MyServer" {
		t.Fatalf("expected 'MyServer', got '%s'", result)
	}
}

func TestExtractNameFromFragment_NoFragment(t *testing.T) {
	u, _ := url.Parse("vmess://abc@host:443")
	result := extractNameFromFragment(u)
	if result != "" {
		t.Fatalf("expected empty string, got '%s'", result)
	}
}

func TestExtractNameFromFragment_EncodedFragment(t *testing.T) {
	u, _ := url.Parse("vmess://abc@host:443#%E4%B8%AD%E6%96%87")
	result := extractNameFromFragment(u)
	if result != "中文" {
		t.Fatalf("expected '中文', got '%s'", result)
	}
}

// ==================== extractNameFromLink ====================

func TestExtractNameFromLink_WithHash(t *testing.T) {
	result := extractNameFromLink("trojan://pass@host:443#TestName")
	if result != "TestName" {
		t.Fatalf("expected 'TestName', got '%s'", result)
	}
}

func TestExtractNameFromLink_NoHash(t *testing.T) {
	result := extractNameFromLink("trojan://pass@host:443")
	if result != "" {
		t.Fatalf("expected empty string, got '%s'", result)
	}
}

func TestExtractNameFromLink_EncodedName(t *testing.T) {
	result := extractNameFromLink("trojan://pass@host:443#%E6%B5%8B%E8%AF%95")
	if result != "测试" {
		t.Fatalf("expected '测试', got '%s'", result)
	}
}

// ==================== removeNameSuffix ====================

func TestRemoveNameSuffix_WithHash(t *testing.T) {
	result := removeNameSuffix("trojan://pass@host:443#TestName")
	expected := "trojan://pass@host:443"
	if result != expected {
		t.Fatalf("expected '%s', got '%s'", expected, result)
	}
}

func TestRemoveNameSuffix_NoHash(t *testing.T) {
	input := "trojan://pass@host:443"
	result := removeNameSuffix(input)
	if result != input {
		t.Fatalf("expected '%s', got '%s'", input, result)
	}
}

func TestRemoveNameSuffix_MultipleHashes(t *testing.T) {
	// Should only remove the last #fragment
	result := removeNameSuffix("trojan://pass@host:443?param=a#b#name")
	expected := "trojan://pass@host:443?param=a#b"
	if result != expected {
		t.Fatalf("expected '%s', got '%s'", expected, result)
	}
}

// ==================== parseQueryString ====================

func TestParseQueryString_Simple(t *testing.T) {
	result := parseQueryString("key1=value1&key2=value2")
	if result["key1"] != "value1" {
		t.Fatalf("expected key1=value1, got %s", result["key1"])
	}
	if result["key2"] != "value2" {
		t.Fatalf("expected key2=value2, got %s", result["key2"])
	}
}

func TestParseQueryString_Empty(t *testing.T) {
	result := parseQueryString("")
	if len(result) != 0 {
		t.Fatalf("expected empty map, got %d entries", len(result))
	}
}

func TestParseQueryString_ValueWithEquals(t *testing.T) {
	result := parseQueryString("key=a=b")
	if result["key"] != "a=b" {
		t.Fatalf("expected 'a=b', got '%s'", result["key"])
	}
}

func TestParseQueryString_NoValue(t *testing.T) {
	result := parseQueryString("key=")
	if result["key"] != "" {
		t.Fatalf("expected empty value, got '%s'", result["key"])
	}
}
