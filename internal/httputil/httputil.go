// Package httputil provides shared helpers for HTTP handlers,
// eliminating repetitive error-handling and parsing boilerplate.
package httputil

import (
	"log/slog"
	"net/http"
	"strconv"
)

// RenderFunc is the signature used by all handlers to render templates.
type RenderFunc func(w http.ResponseWriter, r *http.Request, name string, data map[string]any)

// ServerError logs the error and sends a 500 response.
func ServerError(w http.ResponseWriter, msg string, err error) {
	slog.Error(msg, "error", err)
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

// BadRequest sends a 400 response with the given message.
func BadRequest(w http.ResponseWriter, msg string) {
	http.Error(w, msg, http.StatusBadRequest)
}

// PathInt64 parses a path parameter as int64, returning an error on failure.
func PathInt64(r *http.Request, name string) (int64, error) {
	return strconv.ParseInt(r.PathValue(name), 10, 64)
}

// RequireForm parses the request form, sending 400 on failure.
// Returns true if parsing succeeded.
func RequireForm(w http.ResponseWriter, r *http.Request) bool {
	if err := r.ParseForm(); err != nil {
		BadRequest(w, "Bad Request")
		return false
	}
	return true
}
