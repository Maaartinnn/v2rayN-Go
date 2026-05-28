package coredef

import (
	"io"
	"log/slog"
)

// InitLogger 初始化全局 slog logger，替代标准库 log。
// level 支持 "debug"、"info"（默认）、"warn"、"error"。
func InitLogger(level string, w io.Writer) {
	var slogLevel slog.Level
	switch level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}
	handler := slog.NewTextHandler(w, &slog.HandlerOptions{Level: slogLevel})
	slog.SetDefault(slog.New(handler))
}
