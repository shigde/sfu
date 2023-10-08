package media

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/shigde/sfu/internal/stream"
	"golang.org/x/exp/slog"
)

var (
	errSpaceRequestIdNotFound  = errors.New("reading space id from request")
	errSpaceNotFound           = errors.New("reading space from manager")
	errStreamRequestIdNotFound = errors.New("reading stream id from request")
	errStreamNotFound          = errors.New("reading stream from manager")
)

func getLiveStream(r *http.Request, manager spaceGetCreator) (*stream.LiveStream, *stream.Space, error) {
	space, err := getSpace(r, manager)
	if err != nil {
		return nil, nil, err
	}

	id, ok := mux.Vars(r)["id"]
	if !ok {
		return nil, nil, errStreamRequestIdNotFound
	}

	streamResource, err := space.LiveStreamRepo.FindByUuid(r.Context(), id)
	if err != nil && errors.Is(err, stream.ErrStreamNotFound) {
		return nil, nil, errors.Join(err, errStreamNotFound)
	}

	if err != nil {
		return nil, nil, err
	}

	return streamResource, space, nil
}

func getSpaceId(r *http.Request) (string, error) {
	spaceId, ok := mux.Vars(r)["space"]
	if !ok {
		return "", errSpaceRequestIdNotFound
	}
	return spaceId, nil
}

func getSpace(r *http.Request, manager spaceGetCreator) (*stream.Space, error) {
	spaceId, err := getSpaceId(r)
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
	spaceId, err := getSpaceId(r)
	if err != nil {
		return nil, err
	}
	space, err := manager.GetOrCreateSpace(r.Context(), spaceId)
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
	httpError(w, "error get media", http.StatusInternalServerError, err)
}

func httpError(w http.ResponseWriter, errResponse string, code int, err error) {
	slog.Error(fmt.Sprintf("HTTP: %s", errResponse), "code", code, "err", err)
	http.Error(w, errResponse, code)
}
