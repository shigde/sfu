package media

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/shigde/sfu/pkg/config"
	"github.com/shigde/sfu/pkg/engine"
)

type Server struct {
	server *http.Server
	config *config.ServerConfig
}

func NewServer(config *config.ServerConfig) *Server {
	repository := engine.NewRtpStreamRepository()
	return &Server{
		server: &http.Server{Addr: ":8080", Handler: newRouter(config.AuthConfig, repository)},
		config: config,
	}
}

func (s *Server) Serve() error {
	if err := s.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listening and serve: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shuting down http server: %w", err)
	}
	return nil
}
