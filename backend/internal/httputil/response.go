package httputil

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
	}
}

func ValidationError(w http.ResponseWriter, fields map[string]string) {
	JSON(w, http.StatusBadRequest, map[string]any{
		"error":  "validation_failed",
		"fields": fields,
	})
}

func ErrorResponse(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{
		"error": message,
	})
}

type AppError interface {
	error
	HTTPStatus() int
	HTTPMessage() string
}

func WriteError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(AppError); ok {
		ErrorResponse(w, appErr.HTTPStatus(), appErr.HTTPMessage())
		return
	}
	slog.Error("unhandled error", "error", err)
	ErrorResponse(w, http.StatusInternalServerError, "internal server error")
}
