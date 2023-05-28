package media

import (
	"encoding/json"
	"net/http"

	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/pkg/auth"
	"github.com/shigde/sfu/pkg/engine"
)

var (
	api *webrtc.API
)

type whipOffer struct {
	RoomId   string                    `json:"roomId"`
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

		user, ok := r.Context().Value(auth.PrincipalContextKey).(auth.Principal)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if answer, err := spaceManager.Publish(offer, user); err == nil {
			if err := json.NewEncoder(w).Encode(answer); err != nil {
				http.Error(w, "stream invalid", http.StatusInternalServerError)
			}
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
