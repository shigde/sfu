package routes

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/shigde/sfu/internal/rest"
	"github.com/shigde/sfu/internal/stream"
)

var (
	errSpaceRequestIdNotFound  = errors.New("reading space id from request")
	errSpaceNotFound           = errors.New("reading space from manager")
	errStreamRequestIdNotFound = errors.New("reading stream id from request")
	errStreamNotFound          = errors.New("reading stream from manager")
)

func getLiveStream(r *http.Request, streamService *stream.LiveStreamService) (*stream.LiveStream, string, error) {
	spaceIdentifier, err := getSpaceIdentifier(r)
	if err != nil {
		return nil, spaceIdentifier, err
	}

	streamUUID, ok := mux.Vars(r)["id"]
	if !ok {
		return nil, spaceIdentifier, errStreamRequestIdNotFound
	}

	streamResource, err := streamService.FindByUuidAndSpaceIdentifier(r.Context(), streamUUID, spaceIdentifier)
	if err != nil && errors.Is(err, stream.ErrStreamNotFound) {
		return nil, spaceIdentifier, errors.Join(err, errStreamNotFound)
	}

	if err != nil {
		return nil, spaceIdentifier, err
	}

	return streamResource, spaceIdentifier, nil
}

func getSpaceIdentifier(r *http.Request) (string, error) {
	spaceId, ok := mux.Vars(r)["space"]
	if !ok {
		return "", errSpaceRequestIdNotFound
	}
	return spaceId, nil
}

func getSpace(r *http.Request, manager spaceGetCreator) (*stream.Space, error) {
	spaceId, err := getSpaceIdentifier(r)
	if err != nil {
		return nil, err
	}
	space, err := manager.GetSpace(r.Context(), spaceId)
	if err != nil && errors.Is(err, stream.ErrSpaceNotFound) {
		return nil, errors.Join(err, errSpaceNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("getting space from manager: %w", err)
	}
	return space, nil
}

func getOrCreateSpace(r *http.Request, manager spaceGetCreator) (*stream.Space, error) {
	spaceId, err := getSpaceIdentifier(r)
	if err != nil {
		return nil, err
	}
	space, err := manager.CreateSpace(r.Context(), spaceId)
	if err != nil {
		return nil, err
	}

	return space, nil
}

func handleResourceError(w http.ResponseWriter, err error) {
	if errors.Is(err, errStreamNotFound) || errors.Is(err, errSpaceNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if errors.Is(err, errSpaceRequestIdNotFound) || errors.Is(err, errStreamRequestIdNotFound) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	rest.HttpError(w, "error get media", http.StatusInternalServerError, err)
}
