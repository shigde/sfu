package metric

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/exp/slog"
)

var LobbySessions *LobbySessionMetric

func ExtendRouter(router *mux.Router, config *MetricConfig) error {

	if config.Prometheus.Enable {
		endpoint := config.Prometheus.Endpoint
		httpMetric, err := NewHttpMetric()
		if err != nil {
			return fmt.Errorf("creating http metric setup: %w", err)
		}
		router.Use(GetPrometheusMiddleware(httpMetric))

		LobbySessions, err = NewLobbySessionMetrics()
		if err != nil {
			return fmt.Errorf("creating session metric setup: %w", err)
		}

		router.Path(endpoint).Handler(promhttp.Handler())
	}
	return nil
}

func ServeMetrics(ctx context.Context, config *MetricConfig) error {
	router := mux.NewRouter()
	server := &http.Server{Addr: ":8081", Handler: router}
	if err := ExtendRouter(router, config); err != nil {
		slog.Error("creating metrics", "err", err)
	}

	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("metric server error", "err", err)
		}
	}()
	go func() {
		select {
		case <-ctx.Done():
			slog.Info("shutdown metric server")
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			if err := server.Shutdown(ctx); err != nil {
				slog.Error("shutting down metric server error", "err", err)
			}
			cancel()
		case <-stopped:
			slog.Info("metric server was stopped")
		}
	}()

	return nil
}
