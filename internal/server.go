package internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/shigde/sfu/internal/activitypub"
	"github.com/shigde/sfu/internal/auth"
	"github.com/shigde/sfu/internal/config"
	"github.com/shigde/sfu/internal/lobby"
	"github.com/shigde/sfu/internal/metric"
	"github.com/shigde/sfu/internal/migration"
	"github.com/shigde/sfu/internal/routes"
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
	cfg    *config.SFU
	tp     *trace.TracerProvider
}

func NewServer(ctx context.Context, cfg *config.SFU) (*Server, error) {
	store, err := storage.NewStore(cfg.StorageConfig)
	if err != nil {
		return nil, fmt.Errorf("setup storage %w", err)
	}

	if err := migration.Migrate(cfg.FederationConfig, store); err != nil {
		return nil, fmt.Errorf("build database shema %w", err)
	}

	if cfg.StorageConfig.LoadFixtures {
		if err := migration.LoadFixtures(cfg.FederationConfig, store); err != nil {
			return nil, fmt.Errorf("load fixtures %w", err)
		}
	}

	// RTP lobby
	engine, err := rtp.NewEngine(cfg.RtpConfig)
	if err != nil {
		return nil, fmt.Errorf("creating webrtc engine: %w", err)
	}

	host, _ := url.Parse(cfg.FederationConfig.InstanceUrl.String())
	host.Path = fmt.Sprintf("federation/accounts/%s", cfg.FederationConfig.InstanceUsername)
	lobbyManager := lobby.NewLobbyManager(store, engine, host, cfg.FederationConfig.RegisterToken)

	streamRepo := stream.NewLiveStreamRepository(store)
	spaceRepo := stream.NewSpaceRepository(store)
	liveStreamService := stream.NewLiveStreamService(streamRepo, spaceRepo)
	liveLobbyService := stream.NewLiveLobbyService(store, lobbyManager)

	// Auth provider
	accountRepo := auth.NewAccountRepository(store)
	accountService := auth.NewAccountService(accountRepo, cfg.RegisterToken, cfg.SecurityConfig)

	router := routes.NewRouter(
		cfg.SecurityConfig,
		cfg.RtpConfig,
		accountService,
		liveStreamService,
		liveLobbyService,
	)

	// federation api
	api, err := activitypub.NewApApi(
		cfg.FederationConfig,
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
	if err := metric.ServeMetrics(ctx, cfg.MetricConfig); err != nil {
		return nil, fmt.Errorf("serving metrics: %w", err)
	}

	tp, err := telemetry.NewTracerProvider(ctx, cfg.TelemetryConfig)
	if err != nil {
		return nil, fmt.Errorf("starting telemetry tracer provider: %w", err)
	}

	// mux := http.TimeoutHandler(router, maxRequestTime, "Request Timeout!")
	// start server
	return &Server{
		ctx:    ctx,
		server: &http.Server{Addr: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), Handler: router},
		cfg:    cfg,
		tp:     tp,
	}, nil
}

func (s *Server) Serve() error {
	slog.Info("server Serve() listen", "addr", s.server.Addr)

	if s.cfg.HTTPS {
		if err := s.server.ListenAndServeTLS(s.cfg.Crt, s.cfg.Key); !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listening and serve: %w", err)
		}
		return nil
	}

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
