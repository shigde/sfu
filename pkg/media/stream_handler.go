package media

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/shigde/sfu/pkg/auth"
	"github.com/shigde/sfu/pkg/stream"
)

func getStreamList(manager *stream.SpaceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		space, ok := getSpace(r, manager)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		streams := space.LiveStreamRepo.All()
		if err := json.NewEncoder(w).Encode(streams); err != nil {
			http.Error(w, "reading resources", http.StatusInternalServerError)
		}
	}
}
func getStream(manager *stream.SpaceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		streamResource, _, found := getLiveStream(w, r, manager)
		if !found {
			return
		}

		if err := json.NewEncoder(w).Encode(streamResource); err != nil {
			http.Error(w, "stream invalid", http.StatusInternalServerError)
		}
	}
}

func deleteStream(manager *stream.SpaceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		space, ok := getSpace(r, manager)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		id, ok := mux.Vars(r)["id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if space.LiveStreamRepo.Delete(id) {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func createStream(manager *stream.SpaceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		space, ok := getOrCreateSpace(r, manager)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var liveStream stream.LiveStream
		if err := getStreamResourcePayload(w, r, &liveStream); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		liveStream.User = user.UID
		liveStream.SpaceId = space.Id
		id := space.LiveStreamRepo.Add(&liveStream)
		w.Header().Set("Location", fmt.Sprintf("%s/%s", r.URL.String(), id))
		w.WriteHeader(http.StatusCreated)
	}
}

func updateStream(manager *stream.SpaceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		space, ok := getSpace(r, manager)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var liveStream stream.LiveStream
		if err := getStreamResourcePayload(w, r, &liveStream); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if ok := space.LiveStreamRepo.Update(&liveStream); !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	}
}

func getStreamResourcePayload(w http.ResponseWriter, r *http.Request, liveStream *stream.LiveStream) error {
	dec, err := getPayload(w, r)
	if err != nil {
		return err
	}

	if err := dec.Decode(&liveStream); err != nil {
		return invalidPayload
	}

	return nil
}
