package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/shigde/sfu/pkg/config"
	"github.com/shigde/sfu/pkg/logging"
	"github.com/shigde/sfu/pkg/media"
)

func main() {
	log := logging.NewSlog("development")
	conf, err := config.ParseConfig("config.toml")
	if err != nil {
		log.Error("Parsing Config: %w", err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	server := media.NewMediaServer(conf)

	go func() {
		sig := <-sigs
		log.Info("requesting exit, shutting down by signal: %v", sig)
		server.Shutdown()
	}()

	if err := server.Serve(); err != nil {
		log.Error("Server stopped: %w", err)
	}
}
