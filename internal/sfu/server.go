package sfu

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/shigde/sfu/internal/activitypub"
	"github.com/shigde/sfu/internal/auth"
	"github.com/shigde/sfu/internal/lobby"
	"github.com/shigde/sfu/internal/media"
	"github.com/shigde/sfu/internal/metric"
	"github.com/shigde/sfu/internal/migration"
	"github.com/shigde/sfu/internal/rtp"
	"github.com/shigde/sfu/internal/storage"
	"github.com/shigde/sfu/internal/stream"
	"github.com/shigde/sfu/internal/telemetry"
	"go.opentelemetry.io/otel/sdk/trace"
	"golang.org/x/exp/slog"
)

var maxRequestTime = time.Second * 5

type Server struct {
	ctx    context.Context
	server *http.Server
	config *Config
	tp     *trace.TracerProvider
}

func NewServer(ctx context.Context, config *Config) (*Server, error) {
	store, err := storage.NewStore(config.StorageConfig)
	if err != nil {
		return nil, fmt.Errorf("setup storage %w", err)
	}

	if err := migration.Migrate(config.FederationConfig, store); err != nil {
		return nil, fmt.Errorf("bild database shema %w", err)
	}

	// RTP lobby
	engine, err := rtp.NewEngine(config.RtpConfig)
	if err != nil {
		return nil, fmt.Errorf("creating webrtc engine: %w", err)
	}
	lobbyManager := lobby.NewLobbyManager(engine)

	// @TODO Please do the more simple!!!! To complicated
	repo := stream.NewLiveStreamRepository(store)
	liveStreamService := stream.NewLiveStreamService(repo)
	spaceManager := stream.NewSpaceManager(lobbyManager, store, repo)

	// Auth provider
	accountService, err := auth.NewAccountService(store)
	if err != nil {
		return nil, fmt.Errorf("creating acoount service %w", err)
	}

	router := media.NewRouter(
		config.SecurityConfig,
		config.RtpConfig,
		accountService,
		spaceManager,
	)

	// federation api
	api, err := activitypub.NewApApi(
		config.FederationConfig,
		store,
		liveStreamService,
	)
	if err != nil {
		return nil, fmt.Errorf("creating federation api: %w", err)
	}

	if err := api.BoostrapApi(router); err != nil {
		return nil, fmt.Errorf("boostrapping federation api: %w", err)
	}

	// monitoring
	if err := metric.ExtendRouter(router, config.MetricConfig); err != nil {
		return nil, fmt.Errorf("handling metrics: %w", err)
	}

	tp, err := telemetry.NewTracerProvider(ctx, config.TelemetryConfig)
	if err != nil {
		return nil, fmt.Errorf("starting telemetry tracer provider: %w", err)
	}

	mux := http.TimeoutHandler(router, maxRequestTime, "Request Timeout!")
	// start server
	return &Server{
		ctx:    ctx,
		server: &http.Server{Addr: fmt.Sprintf("%s:%d", config.Host, config.Port), Handler: mux},
		config: config,
		tp:     tp,
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
	if err := s.tp.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutting down tracer provider: %w", err)
	}

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shuting down http server: %w", err)
	}
	return nil
}
