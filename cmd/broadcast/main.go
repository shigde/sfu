package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shigde/sfu/pkg/config"
	"github.com/shigde/sfu/pkg/logging"
	"github.com/shigde/sfu/pkg/media"
)

func main() {
	log := logging.NewSlog()
	conf, err := config.ParseConfig("config.toml")
	if err != nil {
		log.Error("parsing config: %w", err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	server := media.NewMediaServer(conf)

	go func() {
		sig := <-sigs
		log.Info("stopping server, shutting down by signal", "signal", sig)

		shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownRelease()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error("shutting down server gracefully: %v", err)
		}
	}()

	if err := server.Serve(); err != nil {
		log.Error("server stopped: %w", err)
	}
	log.Info("server finished")
}
