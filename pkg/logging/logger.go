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

func NewSlog(appEnv string) *Log {
	var handler slog.Handler = slog.NewTextHandler(os.Stdout)
	if appEnv == "production" {
		handler = slog.NewJSONHandler(os.Stdout)
	}
	return &Log{slog.New(handler)}
}

func (l *Log) logHTTPError(w http.ResponseWriter, err string, code int) {
	l.Error(fmt.Sprintf("HTTP: %s", err), "code", code)
	http.Error(w, err, code)
}
