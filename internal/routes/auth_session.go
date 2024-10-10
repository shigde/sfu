package routes

import (
	"errors"
	"net/http"

	"github.com/shigde/sfu/internal/auth/session"
	"github.com/shigde/sfu/internal/rest"
)

func getUserFromSession(w http.ResponseWriter, r *http.Request) (*session.Principal, error) {
	user, err := session.GetPrincipalFromSession(r)
	if err != nil {
		switch {
		case errors.Is(err, session.ErrNotAuthenticatedSession):
			rest.HttpError(w, "no session", http.StatusForbidden, err)
		case errors.Is(err, session.ErrNoUserSession):
			rest.HttpError(w, "no user session", http.StatusForbidden, err)
		default:
			rest.HttpError(w, "internal error", http.StatusInternalServerError, err)
		}

		return nil, err
	}
	return user, nil
}