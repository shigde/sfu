package routes

import (
	"errors"
	"net/http"

	"github.com/shigde/sfu/internal/auth"
)

func getUserFromSession(w http.ResponseWriter, r *http.Request) (*auth.Principal, error) {
	user, err := auth.GetPrincipalFromSession(r)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrNotAuthenticatedSession):
			httpError(w, "no session", http.StatusForbidden, err)
		case errors.Is(err, auth.ErrNoUserSession):
			httpError(w, "no user session", http.StatusForbidden, err)
		default:
			httpError(w, "internal error", http.StatusInternalServerError, err)
		}

		return nil, err
	}
	return user, nil
}
