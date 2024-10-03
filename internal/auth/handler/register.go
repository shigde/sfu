package handler

import (
	"net/http"

	"github.com/shigde/sfu/internal/auth"
	"github.com/shigde/sfu/internal/rest"
)

func Register(accountService *auth.AccountService) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		account, err := getJsonRegisterPayload(w, r)
		if err != nil {
			rest.HttpError(w, "", http.StatusBadRequest, err)
			return
		}

		err = accountService.CreateAccount(r.Context(), account)
		if err != nil {
			rest.HttpError(w, "registration error", http.StatusBadRequest, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func getJsonRegisterPayload(w http.ResponseWriter, r *http.Request) (*auth.Account, error) {
	dec, err := rest.GetJsonPayload(w, r)
	if err != nil {
		return nil, err
	}
	var account auth.Account
	if err := dec.Decode(&account); err != nil {
		return nil, rest.InvalidPayload
	}
	return &account, nil
}
