package rest

import (
	"fmt"
	"net/http"

	"golang.org/x/exp/slog"
)

func HttpError(w http.ResponseWriter, errResponse string, code int, err error) {
	slog.Error(fmt.Sprintf("HTTP: %s", errResponse), "code", code, "err", err)
	http.Error(w, errResponse, code)
}
