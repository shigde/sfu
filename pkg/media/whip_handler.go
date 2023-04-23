package media

import (
	"net/http"

	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/pkg/engine"
)

var (
	api *webrtc.API
)

type offer struct {
}

func whip(repository *StreamRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var offer offer
		if err := getOfferPayload(w, r, &offer); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		_, err := engine.NewConnection()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	}
}

func getOfferPayload(w http.ResponseWriter, r *http.Request, offer *offer) error {
	dec, err := getPayload(w, r)
	if err != nil {
		return err
	}

	if err := dec.Decode(&offer); err != nil {
		return invalidPayload
	}

	return nil
}
