package media

import (
	"encoding/json"
	"net/http"

	"github.com/shigde/sfu/pkg/auth"
	"github.com/shigde/sfu/pkg/stream"
)

func whip(spaceManager *stream.SpaceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		liveStream, space, err := getLiveStream(r, spaceManager)
		if err != nil {
			handleResourceError(w, err)
			return
		}

		var offer Offer
		if err := getOfferPayload(w, r, &offer); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		user, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		answer, err := space.EnterLobby(&offer.Sdp, liveStream, user.UID, "role")
		if err != nil {
			httpError(w, "error build whip", http.StatusInternalServerError, err)
			return
		}

		if err := json.NewEncoder(w).Encode(answer); err != nil {
			httpError(w, "error build whip", http.StatusInternalServerError, err)
		}
	}
}

func getOfferPayload(w http.ResponseWriter, r *http.Request, offer *Offer) error {
	dec, err := getPayload(w, r)
	if err != nil {
		return err
	}

	if err := dec.Decode(&offer); err != nil {
		return invalidPayload
	}

	return nil
}
