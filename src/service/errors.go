package service

import "fmt"

// ErrNotFound 资源不存在（→ HTTP 404）
type ErrNotFound struct {
	Msg string
	Err error
}

func (e *ErrNotFound) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Err)
	}
	return e.Msg
}

func (e *ErrNotFound) Unwrap() error { return e.Err }

// NewNotFound 创建资源不存在错误
func NewNotFound(msg string, err error) error {
	return &ErrNotFound{Msg: msg, Err: err}
}

// ErrValidation 参数校验失败（→ HTTP 400）
type ErrValidation struct {
	Msg string
	Err error
}

func (e *ErrValidation) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Err)
	}
	return e.Msg
}

func (e *ErrValidation) Unwrap() error { return e.Err }

// NewValidation 创建校验错误
func NewValidation(msg string, err error) error {
	return &ErrValidation{Msg: msg, Err: err}
}

// ErrConflict 业务约束冲突（→ HTTP 409）
type ErrConflict struct {
	Msg string
	Err error
}

func (e *ErrConflict) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Err)
	}
	return e.Msg
}

func (e *ErrConflict) Unwrap() error { return e.Err }

// NewConflict 创建冲突错误
func NewConflict(msg string, err error) error {
	return &ErrConflict{Msg: msg, Err: err}
}
