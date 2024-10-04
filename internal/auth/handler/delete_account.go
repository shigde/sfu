package handler

import (
	"encoding/json"
	"net/http"

	"github.com/shigde/sfu/internal/auth/account"
	"github.com/shigde/sfu/internal/rest"
)

func DeleteAccount(accountService *account.AccountService) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		user, err := getJsonAuthPayload(w, r)
		if err != nil {
			rest.HttpError(w, "", http.StatusBadRequest, err)
			return
		}

		token, err := accountService.GetAuthToken(r.Context(), user)
		if err != nil {
			rest.HttpError(w, "error reading stream list", http.StatusNotFound, err)
			return
		}
		if err := json.NewEncoder(w).Encode(token); err != nil {
			rest.HttpError(w, "error reading stream list", http.StatusInternalServerError, err)
		}
	}
}
