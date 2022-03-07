package middleware

import (
	"context"
	"github.com/shigde/sfu/pkg/auth"
	"github.com/shigde/sfu/pkg/logger"
	"net/http"
)

var log = logger.New()

type contextKey string

const PrincipalContextKey = contextKey("principal")

func AuthMiddleware(ac *auth.AuthConfig, f http.HandlerFunc) http.HandlerFunc {
	log.V(1).Info("Authentication Middleware Is Active.")
	return func(w http.ResponseWriter, r *http.Request) {
		log.V(2).Info("Get New Client Request")
		token, tok := r.URL.Query()["bearer"]

		if tok && len(token) == 1 {
			principal, err := auth.ValidateToken(token[0], ac.GetJwt())
			if err != nil {
				log.V(2).Info("Invalid token")
				http.Error(w, "Forbidden", http.StatusForbidden)

			} else {
				log.V(2).Info("User With Authenticated Token")
				ctx := context.WithValue(r.Context(), PrincipalContextKey, principal)
				f(w, r.WithContext(ctx))
			}

		} else {
			log.V(2).Info("No Authentication Header")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Please login or provide authentication header"))
		}
	}
}