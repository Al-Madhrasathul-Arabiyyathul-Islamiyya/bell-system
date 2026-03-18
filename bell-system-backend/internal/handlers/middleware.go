package handlers

import (
	"context"
	"net/http"
	"strings"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
)

// contextKey is a private type for context keys in this package.
type contextKey string

const (
	// UserContextKey is the key used to store the authenticated user claims in the request context.
	UserContextKey contextKey = "user"
)

// AuthMiddleware returns middleware that validates JWT tokens from the Authorization header.
func AuthMiddleware(tokenSvc TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, http.StatusUnauthorized, "unauthorized", "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				writeError(w, http.StatusUnauthorized, "unauthorized", "invalid authorization format")
				return
			}

			claims, err := tokenSvc.ValidateToken(parts[1])
			if err != nil {
				writeError(w, http.StatusUnauthorized, "unauthorized", "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole returns middleware that restricts access to users with the specified roles.
func RequireRole(roles ...models.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetUserClaims(r.Context())
			if claims == nil {
				writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
				return
			}

			for _, role := range roles {
				if claims.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			writeError(w, http.StatusForbidden, "forbidden", "insufficient permissions")
		})
	}
}

// GetUserClaims extracts the TokenClaims from the request context.
func GetUserClaims(ctx context.Context) *TokenClaims {
	claims, _ := ctx.Value(UserContextKey).(*TokenClaims)
	return claims
}
