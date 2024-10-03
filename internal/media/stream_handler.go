package media

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/shigde/sfu/internal/auth"
	http2 "github.com/shigde/sfu/internal/http"
	"github.com/shigde/sfu/internal/stream"
)

func getStreamList(streamService *stream.LiveStreamService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		spaceIdentifier, err := getSpaceIdentifier(r)
		if err != nil {
			handleResourceError(w, err)
			return
		}
		streams, err := streamService.AllBySpaceIdentifier(r.Context(), spaceIdentifier)
		if err != nil {
			httpError(w, "error reading stream list", http.StatusInternalServerError, err)
			return
		}

		for i, streamResource := range streams {
			if streamResource.Video != nil {
				streams[i].Title = streamResource.Video.Name
			}
		}

		if err := json.NewEncoder(w).Encode(streams); err != nil {
			httpError(w, "error reading stream list", http.StatusInternalServerError, err)
		}
	}
}
func getStream(streamService *stream.LiveStreamService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		streamResource, _, err := getLiveStream(r, streamService)
		if err != nil {
			handleResourceError(w, err)
			return
		}

		if streamResource.Video != nil {
			streamResource.Title = streamResource.Video.Name
		}

		if err := json.NewEncoder(w).Encode(streamResource); err != nil {
			httpError(w, "stream invalid", http.StatusInternalServerError, err)
		}
	}
}

func deleteStream(streamService *stream.LiveStreamService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		user, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		id, ok := mux.Vars(r)["id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := streamService.Delete(r.Context(), id, user.UUID); err != nil {
			httpError(w, "error delete stream", http.StatusInternalServerError, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func createStream(streamService *stream.LiveStreamService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		user, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		spaceIdentifier, err := getSpaceIdentifier(r)
		if err != nil {
			handleResourceError(w, err)
			return
		}

		var liveStream stream.LiveStream
		if err := getStreamResourcePayload(w, r, &liveStream); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		id, err := streamService.CreateStream(r.Context(), &liveStream, spaceIdentifier, user.UUID)
		if err != nil {
			httpError(w, "error create stream", http.StatusInternalServerError, err)
			return
		}
		w.Header().Set("Location", fmt.Sprintf("%s/%s", r.URL.String(), id))
		w.WriteHeader(http.StatusCreated)
	}
}

func updateStream(streamService *stream.LiveStreamService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		user, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		spaceIdentifier, err := getSpaceIdentifier(r)
		if err != nil {
			handleResourceError(w, err)
			return
		}

		var liveStream stream.LiveStream
		if err := getStreamResourcePayload(w, r, &liveStream); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := streamService.UpdateStream(r.Context(), &liveStream, spaceIdentifier, user.UUID); err != nil {
			httpError(w, "error update stream", http.StatusInternalServerError, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func getStreamResourcePayload(w http.ResponseWriter, r *http.Request, liveStream *stream.LiveStream) error {
	dec, err := http2.GetJsonPayload(w, r)
	if err != nil {
		return err
	}

	if err := dec.Decode(&liveStream); err != nil {
		return http2.InvalidPayload
	}

	return nil
}
