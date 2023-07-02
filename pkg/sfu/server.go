package sfu

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shigde/sfu/pkg/lobby"
	"github.com/shigde/sfu/pkg/media"
	"github.com/shigde/sfu/pkg/metric"
	"github.com/shigde/sfu/pkg/storage"
	"github.com/shigde/sfu/pkg/stream"
	"golang.org/x/exp/slog"
)

type Server struct {
	server *http.Server
	config *Config
}

func NewServer(config *Config) (*Server, error) {
	// RTP lobby
	lobbyManager := lobby.NewLobbyManager()

	// Live streams and space
	store, err := storage.NewStore(config.StorageConfig)
	if err != nil {
		return nil, fmt.Errorf("setup storage %w", err)
	}
	spaceManager, err := stream.NewSpaceManager(lobbyManager, store)
	if err != nil {
		return nil, fmt.Errorf("setup space manager %w", err)
	}

	// api endpoints
	router := media.NewRouter(config.AuthConfig, spaceManager)

	// monitoring
	if config.MetricConfig.Prometheus.Enable {
		m, err := metric.NewMetric(config.MetricConfig)
		if err != nil {
			return nil, fmt.Errorf("creating metric setup: %w", err)
		}
		router.Use(metric.GetPrometheusMiddleware(m))
		router.Path(m.Endpoint).Handler(promhttp.Handler())
	}

	// start server
	return &Server{
		server: &http.Server{Addr: ":8080", Handler: router},
		config: config,
	}, nil
}

func (s *Server) Serve() error {
	slog.Info("server Serve() listen", "addr", s.server.Addr)
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
