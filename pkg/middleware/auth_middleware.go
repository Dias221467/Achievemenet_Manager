package middleware

import (
	"context"
	"net/http"
	"strings"

	jwtutil "github.com/Dias221467/Achievemenet_Manager/pkg/jwt"
	"github.com/Dias221467/Achievemenet_Manager/pkg/logger"
)

// Context key for storing user info
type contextKey string

const UserContextKey contextKey = "user"

// AuthMiddleware validates JWT tokens from incoming requests
func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
				return
			}

			// Expect "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Invalid Authorization format", http.StatusUnauthorized)
				return
			}

			// Validate token
			claims, err := jwtutil.ValidateToken(parts[1], secret)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Store user info in context and pass it to the next handler
			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole enforces that the user has a specific role (e.g., "admin")
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetUserFromContext(r.Context())
			if claims == nil || claims.Role != role {
				logger.Log.Warnf("Access denied. Required role: %s, got: %s", role, claims.Role)
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext extracts user info from the request context
func GetUserFromContext(ctx context.Context) *jwtutil.Claims {
	claims, ok := ctx.Value(UserContextKey).(*jwtutil.Claims)
	if !ok {
		return nil
	}
	return claims
}
