package handler

import (
	"encoding/json"
	"net/http"

	"github.com/shigde/sfu/internal/auth/account"
	"github.com/shigde/sfu/internal/rest"
	"github.com/shigde/sfu/pkg/authentication"
)

func Authentication(accountService *account.AccountService) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		user, err := getJsonAuthPayload(w, r)
		if err != nil {
			rest.HttpError(w, "", http.StatusBadRequest, err)
			return
		}

		token, err := accountService.GetAuthToken(r.Context(), user)
		if err != nil {
			rest.HttpError(w, "error authentication", http.StatusNotFound, err)
			return
		}
		if err := json.NewEncoder(w).Encode(token); err != nil {
			rest.HttpError(w, "error auth response", http.StatusInternalServerError, err)
		}
	}
}

func getJsonAuthPayload(w http.ResponseWriter, r *http.Request) (*authentication.ClientUser, error) {
	dec, err := rest.GetJsonPayload(w, r)
	if err != nil {
		return nil, err
	}
	var user authentication.ClientUser
	if err := dec.Decode(&user); err != nil {
		return nil, rest.InvalidPayload
	}
	return &user, nil
}
