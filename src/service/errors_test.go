package service

import (
	"errors"
	"fmt"
	"testing"
)

// ==================== ErrNotFound ====================

func TestErrNotFound_MsgOnly(t *testing.T) {
	err := NewNotFound("profile not found", nil)
	if err.Error() != "profile not found" {
		t.Fatalf("expected 'profile not found', got '%s'", err.Error())
	}
}

func TestErrNotFound_WithCause(t *testing.T) {
	cause := fmt.Errorf("record missing")
	err := NewNotFound("profile not found", cause)
	expected := "profile not found: record missing"
	if err.Error() != expected {
		t.Fatalf("expected '%s', got '%s'", expected, err.Error())
	}
}

func TestErrNotFound_Unwrap_Nil(t *testing.T) {
	err := &ErrNotFound{Msg: "test", Err: nil}
	if err.Unwrap() != nil {
		t.Fatal("expected nil Unwrap")
	}
}

func TestErrNotFound_Unwrap_WithCause(t *testing.T) {
	cause := fmt.Errorf("inner error")
	err := NewNotFound("test", cause)
	if err.(interface{ Unwrap() error }).Unwrap() != cause {
		t.Fatal("Unwrap should return the inner cause")
	}
}

func TestErrNotFound_ErrorsAs(t *testing.T) {
	err := NewNotFound("test", nil)
	var target *ErrNotFound
	if !errors.As(err, &target) {
		t.Fatal("errors.As should match *ErrNotFound")
	}
	if target.Msg != "test" {
		t.Fatalf("expected Msg 'test', got '%s'", target.Msg)
	}
}

// ==================== ErrValidation ====================

func TestErrValidation_MsgOnly(t *testing.T) {
	err := NewValidation("invalid port", nil)
	if err.Error() != "invalid port" {
		t.Fatalf("expected 'invalid port', got '%s'", err.Error())
	}
}

func TestErrValidation_WithCause(t *testing.T) {
	cause := fmt.Errorf("port out of range")
	err := NewValidation("invalid port", cause)
	expected := "invalid port: port out of range"
	if err.Error() != expected {
		t.Fatalf("expected '%s', got '%s'", expected, err.Error())
	}
}

func TestErrValidation_Unwrap(t *testing.T) {
	cause := fmt.Errorf("inner")
	err := NewValidation("test", cause)
	if err.(interface{ Unwrap() error }).Unwrap() != cause {
		t.Fatal("Unwrap should return the inner cause")
	}
}

func TestErrValidation_ErrorsAs(t *testing.T) {
	err := NewValidation("test", nil)
	var target *ErrValidation
	if !errors.As(err, &target) {
		t.Fatal("errors.As should match *ErrValidation")
	}
}

// ==================== ErrConflict ====================

func TestErrConflict_MsgOnly(t *testing.T) {
	err := NewConflict("duplicate name", nil)
	if err.Error() != "duplicate name" {
		t.Fatalf("expected 'duplicate name', got '%s'", err.Error())
	}
}

func TestErrConflict_WithCause(t *testing.T) {
	cause := fmt.Errorf("unique constraint violation")
	err := NewConflict("duplicate name", cause)
	expected := "duplicate name: unique constraint violation"
	if err.Error() != expected {
		t.Fatalf("expected '%s', got '%s'", expected, err.Error())
	}
}

func TestErrConflict_Unwrap(t *testing.T) {
	cause := fmt.Errorf("inner")
	err := NewConflict("test", cause)
	if err.(interface{ Unwrap() error }).Unwrap() != cause {
		t.Fatal("Unwrap should return the inner cause")
	}
}

func TestErrConflict_ErrorsAs(t *testing.T) {
	err := NewConflict("test", nil)
	var target *ErrConflict
	if !errors.As(err, &target) {
		t.Fatal("errors.As should match *ErrConflict")
	}
}

// ==================== Cross-type errors.As ====================

func TestErrorsAs_NotCrossType(t *testing.T) {
	notFound := NewNotFound("test", nil)
	var validationTarget *ErrValidation
	if errors.As(notFound, &validationTarget) {
		t.Fatal("ErrNotFound should not match *ErrValidation")
	}
	var conflictTarget *ErrConflict
	if errors.As(notFound, &conflictTarget) {
		t.Fatal("ErrNotFound should not match *ErrConflict")
	}
}
