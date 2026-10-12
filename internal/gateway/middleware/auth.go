package middleware

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"github.com/artlink52/marketplace/internal/gateway/lib"
	"github.com/artlink52/marketplace/internal/gateway/render"
)

type contextKey string

const claimsKey contextKey = "claims"

func Auth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			res := render.New(w)

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				res.Error(http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				res.Error(http.StatusUnauthorized, "invalid authorization header format")
				return
			}

			claims, err := lib.ParseToken(parts[1], secret)
			if err != nil {
				res.Error(http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				render.New(w).Error(http.StatusUnauthorized, "unauthorized")
				return
			}

			if !slices.Contains(roles, claims.Role) {
				render.New(w).Error(http.StatusForbidden, "forbidden")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func ClaimsFromContext(ctx context.Context) (lib.Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(lib.Claims)
	return claims, ok
}
