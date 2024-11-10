package handler

import (
	"net/http"

	"github.com/shigde/sfu/internal/auth/account"
	"github.com/shigde/sfu/internal/rest"
	"github.com/shigde/sfu/pkg/authentication"
	"golang.org/x/exp/slog"
)

func SendForgotPasswordEmail(accountService *account.AccountService) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		email, err := getForgotPasswordEmailPayload(w, r)
		if err != nil {
			rest.HttpError(w, "", http.StatusBadRequest, err)
			return
		}

		if err = accountService.CreateForgotPasswordToken(r.Context(), email); err != nil {
			slog.Error("auth.SendForgotPasswordEmail:", "err", err)
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func UpdateForgotPassword(accountService *account.AccountService) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		pass, err := getForgotPasswordPayload(w, r)
		if err != nil {
			rest.HttpError(w, "", http.StatusBadRequest, err)
			return
		}

		if err := accountService.UpdatePasswordByToken(r.Context(), pass.Token, pass.Password); err != nil {
			rest.HttpError(w, "error delete account", http.StatusBadRequest, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func getForgotPasswordEmailPayload(w http.ResponseWriter, r *http.Request) (string, error) {
	dec, err := rest.GetJsonPayload(w, r)
	if err != nil {
		return "", err
	}
	var pass authentication.ForgotPasswordEmail
	if err := dec.Decode(&pass); err != nil {
		return "", rest.InvalidPayload
	}

	return pass.Email, nil
}

func getForgotPasswordPayload(w http.ResponseWriter, r *http.Request) (*authentication.ForgotPassword, error) {
	dec, err := rest.GetJsonPayload(w, r)
	if err != nil {
		return nil, err
	}
	var pass authentication.ForgotPassword
	if err := dec.Decode(&pass); err != nil {
		return nil, rest.InvalidPayload
	}

	return &pass, nil
}
