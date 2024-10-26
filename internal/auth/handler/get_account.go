package handler

import (
	"encoding/json"
	"net/http"

	"github.com/shigde/sfu/internal/auth/account"
	"github.com/shigde/sfu/internal/auth/session"
	"github.com/shigde/sfu/internal/rest"
)

func GetAccount(accountService *account.AccountService) http.HandlerFunc {

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

		account, err := accountService.GetAccount(r.Context(), &userUuid)
		if err != nil {
			rest.HttpError(w, "error reading account", http.StatusNotFound, err)
			return
		}
		if err := json.NewEncoder(w).Encode(account); err != nil {
			rest.HttpError(w, "error reading account", http.StatusInternalServerError, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
