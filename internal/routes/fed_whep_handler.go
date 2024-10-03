package routes

import (
	"net/http"

	"github.com/shigde/sfu/internal/stream"
	"go.opentelemetry.io/otel"
)

func fedWhep(streamService *stream.LiveStreamService, liveService *stream.LiveLobbyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, span := otel.Tracer(tracerName).Start(r.Context(), "api: fed_whep_create")
		defer span.End()
	}
}
