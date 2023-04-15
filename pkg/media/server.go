package media

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/shigde/sfu/pkg/auth"
	"github.com/shigde/sfu/pkg/config"
)

type MediaServer struct {
	server *http.Server
	config *config.ServerConfig
	router *mux.Router
}

func NewMediaServer(config *config.ServerConfig) *MediaServer {
	router := mux.NewRouter()
	return &MediaServer{
		server: &http.Server{Addr: ":8080", Handler: router},
		config: config,
		router: router,
	}
}

func (s *MediaServer) Serve() error {
	repository := newStreamRepository()

	s.router.HandleFunc("/streams", auth.HttpMiddleware(s.config.AuthConfig, getStreamList(*repository))).Methods("GET")
	s.router.HandleFunc("/stream", auth.HttpMiddleware(s.config.AuthConfig, createStream(*repository))).Methods("POST")
	s.router.HandleFunc("/stream/{id}", auth.HttpMiddleware(s.config.AuthConfig, getStream(*repository))).Methods("GET")
	s.router.HandleFunc("/stream", auth.HttpMiddleware(s.config.AuthConfig, updateStream(*repository))).Methods("PUT")
	s.router.HandleFunc("/stream/{id}", auth.HttpMiddleware(s.config.AuthConfig, deleteStream(*repository))).Methods("DELETE")

	if err := s.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listening and serve: %w", err)
	}
	return nil
}

func (s *MediaServer) Shutdown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shuting down http server: %w", err)
	}
	return nil
}
