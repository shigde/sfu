package media

import (
	"github.com/gorilla/mux"
	"github.com/shigde/sfu/pkg/auth"
	"github.com/shigde/sfu/pkg/stream"
)

func NewRouter(
	config *auth.AuthConfig,
	manager *stream.SpaceManager,
) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/space/{space}/streams", auth.HttpMiddleware(config, getStreamList(manager))).Methods("GET")
	router.HandleFunc("/space/{space}/stream", auth.HttpMiddleware(config, createStream(manager))).Methods("POST")
	router.HandleFunc("/space/{space}/stream/{id}", auth.HttpMiddleware(config, getStream(manager))).Methods("GET")
	router.HandleFunc("/space/{space}/stream", auth.HttpMiddleware(config, updateStream(manager))).Methods("PUT")
	router.HandleFunc("/space/{space}/stream/{id}", auth.HttpMiddleware(config, deleteStream(manager))).Methods("DELETE")
	return router
}
