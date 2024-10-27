package handler

import (
	"net/http"

	"github.com/shigde/sfu/internal/auth/account"
	"github.com/shigde/sfu/internal/auth/session"
	"github.com/shigde/sfu/internal/rest"
)

func DeleteAccount(accountService *account.AccountService) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		user, hasLogin := session.PrincipalFromContext(r.Context())
		if !hasLogin {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		userUuid, err := user.GetUuid()
		if err != nil {
			rest.HttpError(w, "session error", http.StatusInternalServerError, err)
			return
		}

		if err := accountService.DeleteAccount(r.Context(), &userUuid); err != nil {
			rest.HttpError(w, "error delete account", http.StatusBadRequest, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
