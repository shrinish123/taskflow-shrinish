package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/shrinishv/taskflow-shrinish/backend/internal/httputil"
	"github.com/shrinishv/taskflow-shrinish/backend/internal/service"
)

type contextKey string

const userIDKey contextKey = "user_id"

func Auth(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				httputil.ErrorResponse(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				httputil.ErrorResponse(w, http.StatusUnauthorized, "invalid authorization header format")
				return
			}

			userID, err := authService.ValidateToken(parts[1])
			if err != nil {
				httputil.ErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(userIDKey).(string); ok {
		return id
	}
	return ""
}
