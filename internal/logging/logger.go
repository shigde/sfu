package logging

import (
	"fmt"
	"net/http"
	"os"

	"golang.org/x/exp/slog"
)

type Log struct {
	file *os.File
	*slog.Logger
}

type LogConfig struct {
	Logfile string `mapstructure:"logfile"`
	Level   string `mapstructure:"level"`
}

func NewSlog(config *LogConfig) (*Log, error) {
	var (
		env     = os.Getenv("APP_ENV")
		handler slog.Handler
	)

	var file *os.File
	var err error

	switch {
	case config.Logfile != "stdout":
		file, err = os.OpenFile(config.Logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("opening log file: %v", err)
		}
	default:
		file = os.Stdout
	}

	//Log level
	var logLevel slog.Level
	switch config.Level {
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	case "DEBUG":
		logLevel = slog.LevelDebug
	default:
		logLevel = slog.LevelInfo
	}
	opts := slog.HandlerOptions{Level: logLevel}

	// Log type
	switch env {
	case "production":
		handler = opts.NewJSONHandler(file)
	default:
		handler = opts.NewTextHandler(file)
	}

	slog.SetDefault(slog.New(handler))
	return &Log{file, slog.Default()}, nil
}

func (l *Log) HTTPError(w http.ResponseWriter, err string, code int) {
	l.Debug(fmt.Sprintf("HTTP: %s", err), "code", code)
	http.Error(w, err, code)
}
func (l *Log) Close() {
	if l.file != nil {
		_ = l.file.Close()
	}
}

func SetupDebugLogger() {
	opts := slog.HandlerOptions{Level: slog.LevelDebug}
	h := opts.NewTextHandler(os.Stderr)
	slog.SetDefault(slog.New(h))
}
