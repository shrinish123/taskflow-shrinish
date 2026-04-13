package handler

import (
	"net/http"
	"regexp"

	"github.com/go-chi/chi/v5"
)

var uuidRegex = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func IsValidUUID(s string) bool {
	return uuidRegex.MatchString(s)
}

func ValidID(w http.ResponseWriter, r *http.Request, param string) (string, bool) {
	id := chi.URLParam(r, param)
	if !IsValidUUID(id) {
		ValidationError(w, map[string]string{param: "must be a valid UUID"})
		return "", false
	}
	return id, true
}
