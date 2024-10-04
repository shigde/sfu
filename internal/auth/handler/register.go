package handler

import (
	"net/http"

	"github.com/shigde/sfu/internal/auth/account"
	"github.com/shigde/sfu/internal/rest"
)

func Register(accountService *account.AccountService) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		acc, err := getJsonRegisterPayload(w, r)
		if err != nil {
			rest.HttpError(w, "", http.StatusBadRequest, err)
			return
		}

		err = accountService.CreateAccount(r.Context(), acc)
		if err != nil {
			rest.HttpError(w, "registration error", http.StatusBadRequest, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func getJsonRegisterPayload(w http.ResponseWriter, r *http.Request) (*account.Account, error) {
	dec, err := rest.GetJsonPayload(w, r)
	if err != nil {
		return nil, err
	}
	var acc account.Account
	if err := dec.Decode(&acc); err != nil {
		return nil, rest.InvalidPayload
	}

	if acc.Password, err = account.HashPassword(acc.Password); err != nil {
		return nil, err
	}
	return &acc, nil
}
