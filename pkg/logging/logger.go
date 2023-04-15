package logging

import (
	"fmt"
	"net/http"
	"os"

	"golang.org/x/exp/slog"
)

type Log struct {
	*slog.Logger
}

func NewSlog() *Log {
	var (
		env     = os.Getenv("APP_ENV")
		handler slog.Handler
	)
	switch env {
	case "production":
		handler = slog.NewJSONHandler(os.Stdout)
	default:
		handler = slog.NewTextHandler(os.Stdout)
	}
	return &Log{slog.New(handler)}
}

func (l *Log) HTTPError(w http.ResponseWriter, err string, code int) {
	l.Debug(fmt.Sprintf("HTTP: %s", err), "code", code)
	http.Error(w, err, code)
}
