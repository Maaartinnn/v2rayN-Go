package web

import (
	"errors"
	"log/slog"
	"net/http"

	"v2rayn-go/service"
)

// mapServiceError 根据 Service 层错误类型映射为对应 HTTP 状态码。
// 业务错误（400/404/409）直接返回给前端；500 仅返回泛化提示，细节写入 slog。
func mapServiceError(w http.ResponseWriter, err error) {
	if e, ok := errors.AsType[*service.ErrNotFound](err); ok {
		jsonError(w, e.Msg, http.StatusNotFound)
		return
	}
	if e, ok := errors.AsType[*service.ErrValidation](err); ok {
		jsonError(w, e.Msg, http.StatusBadRequest)
		return
	}
	if e, ok := errors.AsType[*service.ErrConflict](err); ok {
		jsonError(w, e.Msg, http.StatusConflict)
		return
	}
	// 500：内部细节写日志，前端仅收到泛化提示
	slog.Error("internal server error", "error", err)
	jsonError(w, "Internal Server Error", http.StatusInternalServerError)
}
