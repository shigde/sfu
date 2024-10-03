package routes

import (
	"encoding/json"
	"net/http"

	"github.com/shigde/sfu/internal/auth"
	http2 "github.com/shigde/sfu/internal/http"
	"github.com/shigde/sfu/pkg/authentication"
)

func getAuthenticationHandler(
	accountService *auth.AccountService,
) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		user, err := getJsonAuthPayload(w, r)
		if err != nil {
			httpError(w, "", http.StatusBadRequest, err)
			return
		}

		token, err := accountService.GetAuthToken(r.Context(), user)
		if err != nil {
			httpError(w, "error reading stream list", http.StatusNotFound, err)
			return
		}
		if err := json.NewEncoder(w).Encode(token); err != nil {
			httpError(w, "error reading stream list", http.StatusInternalServerError, err)
		}

	}
}

func getJsonAuthPayload(w http.ResponseWriter, r *http.Request) (*authentication.User, error) {
	dec, err := http2.GetJsonPayload(w, r)
	if err != nil {
		return nil, err
	}
	var user authentication.User
	if err := dec.Decode(&user); err != nil {
		return nil, http2.InvalidPayload
	}
	return &user, nil
}
