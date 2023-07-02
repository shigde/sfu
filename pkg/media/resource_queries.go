package media

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/shigde/sfu/pkg/stream"
)

func getLiveStream(w http.ResponseWriter, r *http.Request, manager *stream.SpaceManager) (*stream.LiveStream, *stream.Space, bool) {
	space, ok := getSpace(r, manager)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return nil, nil, false
	}

	id, ok := mux.Vars(r)["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return nil, nil, false
	}

	if streamResource, ok := space.LiveStreamRepo.FindById(id); ok {
		return streamResource, space, ok
	}
	w.WriteHeader(http.StatusNotFound)
	return nil, nil, false
}

func getSpaceId(r *http.Request) (string, bool) {
	spaceId, ok := mux.Vars(r)["space"]
	return spaceId, ok
}

func getSpace(r *http.Request, manager *stream.SpaceManager) (*stream.Space, bool) {
	if spaceId, ok := getSpaceId(r); ok {
		space, ok := manager.GetSpace(r.Context(), spaceId)
		return space, ok
	}
	return nil, false
}

func getOrCreateSpace(r *http.Request, manager *stream.SpaceManager) (*stream.Space, bool) {
	if spaceId, ok := getSpaceId(r); ok {
		space := manager.GetOrCreateSpace(r.Context(), spaceId)
		return space, ok
	}
	return nil, false
}
