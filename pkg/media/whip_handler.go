package media

import (
	"encoding/json"
	"net/http"

	"github.com/shigde/sfu/pkg/auth"
)

func whip(spaceManager spaceGetCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/sdp")
		if err := auth.StartSession(w, r); err != nil {
			httpError(w, "error", http.StatusInternalServerError, err)
		}

		liveStream, space, err := getLiveStream(r, spaceManager)
		if err != nil {
			handleResourceError(w, err)
			return
		}

		offer, err := getSdpPayload(w, r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		user, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		answer, err := space.EnterLobby(offer, liveStream, user.UID, "role")
		resourceId := "1234567" // The lobby generate this resource id!

		if err != nil {
			httpError(w, "error build whip", http.StatusInternalServerError, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(answer); err != nil {
			httpError(w, "error build whip", http.StatusInternalServerError, err)
		}
		w.Header().Set("etag", "etag")
		w.Header().Set("Content-Length", "1400")
		w.Header().Set("Location", "resource/"+resourceId)
	}
}
