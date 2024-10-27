package handler

import (
	"net/http"

	"github.com/shigde/sfu/internal/auth/account"
	"github.com/shigde/sfu/internal/rest"
	"github.com/shigde/sfu/pkg/authentication"
	"golang.org/x/exp/slog"
)

func ForgotPassword(accountService *account.AccountService) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		email, err := getEmailForPasswordForgotPayload(w, r)
		if err != nil {
			rest.HttpError(w, "", http.StatusBadRequest, err)
			return
		}

		if err = accountService.CreateForgotPasswordToken(r.Context(), email); err != nil {
			slog.Error("auth.ForgotPassword:", "err", err)
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func getEmailForPasswordForgotPayload(w http.ResponseWriter, r *http.Request) (string, error) {
	dec, err := rest.GetJsonPayload(w, r)
	if err != nil {
		return "", err
	}
	var pass authentication.PasswordForgotEmail
	if err := dec.Decode(&pass); err != nil {
		return "", rest.InvalidPayload
	}

	return pass.Email, nil
}
