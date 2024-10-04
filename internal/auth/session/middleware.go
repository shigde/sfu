package session

import (
	"context"
	"net/http"
	"strings"

	"golang.org/x/exp/slog"
)

type contextKey string

var log = slog.Default()

const principalContextKey = contextKey("principal")

func HttpMiddleware(ac *SecurityConfig, f http.HandlerFunc) http.HandlerFunc {
	slog.Debug("activated authentication http middleware")
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("getting new client request")
		authHeader := r.Header.Get("Authorization")

		if !strings.HasPrefix(authHeader, "Bearer ") {
			log.Warn("checking authorization header bearer prefix failing")
			http.Error(w, "Invalid authentication header", http.StatusBadRequest)
			return
		}

		jwtToken := strings.TrimPrefix(authHeader, "Bearer ")

		principal, err := ValidateToken(jwtToken, ac.JWT)
		if err != nil {
			slog.Warn("validating invalid jwt token: %w", err)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		slog.Debug("authenticating user")
		ctx := withPrincipal(r.Context(), principal)
		f(w, r.WithContext(ctx))
	}
}

func withPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalContextKey, principal)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalContextKey).(Principal)
	return principal, ok
}
