package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shigde/sfu/internal/config"
	"github.com/shigde/sfu/internal/logging"
	"github.com/shigde/sfu/internal/sfu"
)

func main() {
	ctx := context.Background()
	configArg := flag.String("config", "config.toml", "config file: default config.toml")
	flag.Parse()

	fmt.Println("config:", *configArg)

	env := config.ParseEnv()
	conf, err := config.ParseConfig(*configArg, env)
	if err != nil {
		panic(fmt.Errorf("parsing config: %w", err))
		return
	}

	log, err := logging.NewSlog(conf.LogConfig)
	if err != nil {
		panic(fmt.Errorf("setup logger: %w", err))
		return
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	server, err := sfu.NewServer(ctx, conf)
	if err != nil {
		panic(fmt.Errorf("creating new server: %w", err))
	}

	go func() {
		sig := <-sigs
		log.Info("stopping server, shutting down by signal", "signal", sig)

		shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownRelease()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error("shutting down server gracefully: %v", "err", err)
		}
	}()

	if err := server.Serve(); err != nil {
		log.Error("server stopped: ", "err", err)
	}
	log.Info("server finished")
	log.Close()
}
