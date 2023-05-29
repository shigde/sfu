package media

import (
	"encoding/json"
	"net/http"

	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/pkg/auth"
	"github.com/shigde/sfu/pkg/engine"
)

type whipOffer struct {
	SpaceId  string                    `json:"spaceId"`
	StreamId string                    `json:"streamId"`
	Offer    webrtc.SessionDescription `json:"offer"`
}

func whip(spaceManager *engine.SpaceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var offer whipOffer
		if err := getOfferPayload(w, r, &offer); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		user, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if space, ok := spaceManager.GetSpace(""); ok {
			if answer, err := space.Publish(offer, user); err != nil {
				if err := json.NewEncoder(w).Encode(answer); err != nil {
					http.Error(w, "stream invalid", http.StatusInternalServerError)
				}
				return
			}
			http.Error(w, "could not publish stream", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}
}

func getOfferPayload(w http.ResponseWriter, r *http.Request, offer *whipOffer) error {
	dec, err := getPayload(w, r)
	if err != nil {
		return err
	}

	if err := dec.Decode(&offer); err != nil {
		return invalidPayload
	}

	return nil
}
