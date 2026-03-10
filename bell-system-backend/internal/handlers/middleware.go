package handlers

import (
	"context"
	"net/http"

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
			// TODO: implement
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole returns middleware that restricts access to users with the specified roles.
func RequireRole(roles ...models.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: implement
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserClaims extracts the TokenClaims from the request context.
func GetUserClaims(ctx context.Context) *TokenClaims {
	claims, _ := ctx.Value(UserContextKey).(*TokenClaims)
	return claims
}
