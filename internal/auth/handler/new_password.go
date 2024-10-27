package handler

import (
	"net/http"

	"github.com/shigde/sfu/internal/auth/account"
	"github.com/shigde/sfu/internal/auth/session"
	"github.com/shigde/sfu/internal/rest"
	"github.com/shigde/sfu/pkg/authentication"
)

func NewPasswordByLogin(accountService *account.AccountService) http.HandlerFunc {

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

		pass, err := getPassForgetPayload(w, r)
		if err != nil {
			rest.HttpError(w, "", http.StatusBadRequest, err)
			return
		}

		if err := accountService.UpdatePassword(r.Context(), &userUuid, pass.OldPassword, pass.NewPassword); err != nil {
			rest.HttpError(w, "error delete account", http.StatusBadRequest, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func NewPasswordByForgot(accountService *account.AccountService) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		pass, err := getPassForgetWithTokenPayload(w, r)
		if err != nil {
			rest.HttpError(w, "", http.StatusBadRequest, err)
			return
		}

		if err := accountService.UpdatePasswordByToken(r.Context(), pass.Token, pass.OldPassword, pass.NewPassword); err != nil {
			rest.HttpError(w, "error delete account", http.StatusBadRequest, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func getPassForgetPayload(w http.ResponseWriter, r *http.Request) (*authentication.PasswordForgot, error) {
	dec, err := rest.GetJsonPayload(w, r)
	if err != nil {
		return nil, err
	}
	var pass authentication.PasswordForgot
	if err := dec.Decode(&pass); err != nil {
		return nil, rest.InvalidPayload
	}

	return &pass, nil
}

func getPassForgetWithTokenPayload(w http.ResponseWriter, r *http.Request) (*authentication.PasswordForgotWithToken, error) {
	dec, err := rest.GetJsonPayload(w, r)
	if err != nil {
		return nil, err
	}
	var pass authentication.PasswordForgotWithToken
	if err := dec.Decode(&pass); err != nil {
		return nil, rest.InvalidPayload
	}

	return &pass, nil
}
