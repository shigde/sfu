package media

import (
	"github.com/gorilla/mux"
	"github.com/shigde/sfu/pkg/auth"
)

func newRouter(config *auth.AuthConfig, repository *StreamRepository) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/streams", auth.HttpMiddleware(config, getStreamList(repository))).Methods("GET")
	router.HandleFunc("/stream", auth.HttpMiddleware(config, createStream(repository))).Methods("POST")
	router.HandleFunc("/stream/{id}", auth.HttpMiddleware(config, getStream(repository))).Methods("GET")
	router.HandleFunc("/stream", auth.HttpMiddleware(config, updateStream(repository))).Methods("PUT")
	router.HandleFunc("/stream/{id}", auth.HttpMiddleware(config, deleteStream(repository))).Methods("DELETE")
	return router
}
