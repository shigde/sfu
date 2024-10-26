package handler

import (
	"encoding/json"
	"net/http"

	"github.com/shigde/sfu/internal/auth/account"
	"github.com/shigde/sfu/internal/rest"
	"github.com/shigde/sfu/pkg/authentication"
)

func Login(accountService *account.AccountService) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		login, err := getJsonLoginPayload(w, r)
		if err != nil {
			rest.HttpError(w, "", http.StatusBadRequest, err)
			return
		}

		token, err := accountService.GetAuthTokenByLogin(r.Context(), login)
		if err != nil {
			rest.HttpError(w, "error login", http.StatusNotFound, err)
			return
		}

		if err := json.NewEncoder(w).Encode(token); err != nil {
			rest.HttpError(w, "error login response", http.StatusInternalServerError, err)
		}
	}
}

func getJsonLoginPayload(w http.ResponseWriter, r *http.Request) (*authentication.Login, error) {
	dec, err := rest.GetJsonPayload(w, r)
	if err != nil {
		return nil, err
	}
	var login authentication.Login
	if err := dec.Decode(&login); err != nil {
		return nil, rest.InvalidPayload
	}
	return &login, nil
}
