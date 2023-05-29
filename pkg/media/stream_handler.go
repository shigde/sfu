package media

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/shigde/sfu/pkg/engine"
)

func getStreamList(manager *engine.SpaceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		space, ok := getSpace(r, manager)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		streams := space.PublicStreamRepo.All()
		if err := json.NewEncoder(w).Encode(streams); err != nil {
			http.Error(w, "reading resources", http.StatusInternalServerError)
		}
	}
}
func getStream(manager *engine.SpaceManager) http.HandlerFunc {
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

		if streamResource, ok := space.PublicStreamRepo.FindById(id); ok {
			if err := json.NewEncoder(w).Encode(streamResource); err != nil {
				http.Error(w, "stream invalid", http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}
}

func deleteStream(manager *engine.SpaceManager) http.HandlerFunc {
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

		if space.PublicStreamRepo.Delete(id) {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func createStream(manager *engine.SpaceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		space, ok := getOrCreateSpace(r, manager)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var stream engine.RtpStream
		if err := getStreamResourcePayload(w, r, &stream); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		id := space.PublicStreamRepo.Add(&stream)
		w.Header().Set("Location", fmt.Sprintf("%s/%s", r.URL.String(), id))
		w.WriteHeader(http.StatusCreated)
	}
}

func updateStream(manager *engine.SpaceManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		space, ok := getSpace(r, manager)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var stream engine.RtpStream
		if err := getStreamResourcePayload(w, r, &stream); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if ok := space.PublicStreamRepo.Update(&stream); !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	}
}

func getStreamResourcePayload(w http.ResponseWriter, r *http.Request, stream *engine.RtpStream) error {
	dec, err := getPayload(w, r)
	if err != nil {
		return err
	}

	if err := dec.Decode(&stream); err != nil {
		return invalidPayload
	}

	return nil
}

func getSpaceId(r *http.Request) (string, bool) {
	spaceId, ok := mux.Vars(r)["space"]
	return spaceId, ok
}

func getSpace(r *http.Request, manager *engine.SpaceManager) (*engine.Space, bool) {
	if spaceId, ok := getSpaceId(r); ok {
		space, ok := manager.GetSpace(spaceId)
		return space, ok
	}
	return nil, false
}

func getOrCreateSpace(r *http.Request, manager *engine.SpaceManager) (*engine.Space, bool) {
	if spaceId, ok := getSpaceId(r); ok {
		space := manager.GetOrCreateSpace(spaceId)
		return space, ok
	}
	return nil, false
}
