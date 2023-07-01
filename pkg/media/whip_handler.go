package media

import (
	"encoding/json"
	"net/http"

	"github.com/shigde/sfu/pkg/auth"
	"github.com/shigde/sfu/pkg/stream"
)

func whip(spaceManager *stream.SpaceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		liveStream, space, found := getLiveStream(w, r, spaceManager)
		if !found {
			return
		}

		var offer Offer
		if err := getOfferPayload(w, r, &offer); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		user, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if answer, err := space.EnterLobby(&offer.Sdp, liveStream, user.UID, "role"); err == nil {
			if err := json.NewEncoder(w).Encode(answer); err != nil {
				http.Error(w, "stream invalid", http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusForbidden)
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
