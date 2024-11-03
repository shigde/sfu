package handler

import (
	"encoding/json"
	"net/http"

	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/auth/account"
	"github.com/shigde/sfu/internal/auth/session"
	"github.com/shigde/sfu/internal/rest"
	"github.com/shigde/sfu/pkg/authentication"
)

func GetUser(accountService *account.AccountService) http.HandlerFunc {

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

		acc, err := accountService.GetAccount(r.Context(), &userUuid)
		if err != nil {
			rest.HttpError(w, "error reading account", http.StatusNotFound, err)
			return
		}

		userName, hostName := instance.SplitUserId(acc.User)
		if err := json.NewEncoder(w).Encode(&authentication.User{Name: userName, Domain: hostName, Role: account.RoleToString(acc.Role)}); err != nil {
			rest.HttpError(w, "error reading account", http.StatusInternalServerError, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
