package sfu

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shigde/sfu/pkg/engine"
	"github.com/shigde/sfu/pkg/media"
	"github.com/shigde/sfu/pkg/metric"
	"golang.org/x/exp/slog"
)

type Server struct {
	server *http.Server
	config *Config
}

func NewServer(config *Config) (*Server, error) {
	repository := engine.NewRtpStreamRepository()
	router := media.NewRouter(config.AuthConfig, repository)

	// monitoring
	if config.MetricConfig.Prometheus.Enable {
		router.Path("/prometheus").Handler(promhttp.Handler())

		m, err := metric.NewMetric(config.MetricConfig)
		if err != nil {
			return nil, fmt.Errorf("creating metric setup: %w", err)
		}

		router.Use(m.GetPrometheusMiddleware())
	}

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
