package auth

import (
	"context"
	"net/http"
	"strings"

	"golang.org/x/exp/slog"
)

type contextKey string

var log = slog.Default()

const principalContextKey = contextKey("principal")

func HttpMiddleware(ac *AuthConfig, f http.HandlerFunc) http.HandlerFunc {
	log.Debug("activated authentication http middleware")
	return func(w http.ResponseWriter, r *http.Request) {
		r.Context()
		w.Header().Set("Content-Type", "application/json")
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
			log.Warn("validating invalid jwt token: %w", err)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		log.Debug("authenticating user")
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
