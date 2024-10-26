package handler

import (
	"net/http"

	"github.com/shigde/sfu/internal/auth/account"
	"github.com/shigde/sfu/internal/rest"
	"github.com/shigde/sfu/pkg/authentication"
)

func Verification(accountService *account.AccountService) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		token, err := getVerifyTokenPayload(w, r)
		if err != nil {
			rest.HttpError(w, "", http.StatusBadRequest, err)
			return
		}

		if err = accountService.VerifyAccount(r.Context(), token); err != nil {
			rest.HttpError(w, "token verification fails", http.StatusBadRequest, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func getVerifyTokenPayload(w http.ResponseWriter, r *http.Request) (string, error) {
	dec, err := rest.GetJsonPayload(w, r)
	if err != nil {
		return "", err
	}
	var verify authentication.Verify
	if err := dec.Decode(&verify); err != nil {
		return "", rest.InvalidPayload
	}

	return verify.Token, nil
}
