package media

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/auth"
)

func getAuthenticationHandler(
	accountService *auth.AccountService,
) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		var uuid *uuid.UUID
		credential := ""
		_, err := accountService.GetAuthToken(uuid, credential)
		if err != nil {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}
}
