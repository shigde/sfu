package media

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/shigde/sfu/internal/auth"
	"github.com/shigde/sfu/internal/stream"
)

func NewRouter(
	config *auth.AuthConfig,
	spaceManager spaceGetCreator,
) *mux.Router {
	router := mux.NewRouter()
	// Space
	router.HandleFunc("/space/{space}/streams", auth.HttpMiddleware(config, getStreamList(spaceManager))).Methods("GET")
	router.HandleFunc("/space/{space}/stream", auth.HttpMiddleware(config, createStream(spaceManager))).Methods("POST")
	router.HandleFunc("/space/{space}/stream/{id}", auth.HttpMiddleware(config, getStream(spaceManager))).Methods("GET")
	router.HandleFunc("/space/{space}/stream", auth.HttpMiddleware(config, updateStream(spaceManager))).Methods("PUT")
	router.HandleFunc("/space/{space}/stream/{id}", auth.HttpMiddleware(config, deleteStream(spaceManager))).Methods("DELETE")
	// Lobby
	router.HandleFunc("/space/{space}/stream/{id}/whip", auth.HttpMiddleware(config, whip(spaceManager))).Methods("POST")
	return router
}

type spaceGetCreator interface {
	GetSpace(ctx context.Context, id string) (*stream.Space, error)
	GetOrCreateSpace(ctx context.Context, id string) (*stream.Space, error)
}
