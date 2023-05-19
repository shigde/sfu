package logging

import (
	"fmt"
	"log"
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
}

func NewSlog(config *LogConfig) (*Log, error) {
	var (
		env     = os.Getenv("APP_ENV")
		handler slog.Handler
	)

	file, err := os.OpenFile(config.Logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("opening log file: %v", err)
	}

	switch env {
	case "production":
		handler = slog.NewJSONHandler(file)
	default:
		handler = slog.NewTextHandler(file)
	}
	slog.SetDefault(slog.New(handler))
	log.Println("This is a test log entry")
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
