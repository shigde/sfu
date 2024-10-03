package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/shigde/sfu/internal/rest"
	"github.com/shigde/sfu/internal/rtp"
)

func getSettings(config *rtp.RtpConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(config.ICEServer); err != nil {
			rest.HttpError(w, "stream invalid", http.StatusInternalServerError, err)
		}
		w.Header().Set("X-CSRF-Token", csrf.Token(r))
	}
}
