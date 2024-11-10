package handler

import (
	"net/http"

	"github.com/shigde/sfu/internal/auth/account"
	"github.com/shigde/sfu/internal/auth/session"
	"github.com/shigde/sfu/internal/rest"
	"github.com/shigde/sfu/pkg/authentication"
)

func UpdatePassword(accountService *account.AccountService) http.HandlerFunc {

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

		pass, err := getUpdatePasswordPayload(w, r)
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

func getUpdatePasswordPayload(w http.ResponseWriter, r *http.Request) (*authentication.UpdatePassword, error) {
	dec, err := rest.GetJsonPayload(w, r)
	if err != nil {
		return nil, err
	}
	var pass authentication.UpdatePassword
	if err := dec.Decode(&pass); err != nil {
		return nil, rest.InvalidPayload
	}

	return &pass, nil
}
