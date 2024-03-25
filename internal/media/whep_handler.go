package media

import (
	"crypto/md5"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/auth"
	"github.com/shigde/sfu/internal/lobby"
	"github.com/shigde/sfu/internal/stream"
	"github.com/shigde/sfu/internal/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func whep(streamService *stream.LiveStreamService, liveService *stream.LiveLobbyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := otel.Tracer(tracerName).Start(r.Context(), "api: whep_create")
		defer span.End()

		w.Header().Set("Content-Type", "application/sdp")
		user, err := auth.GetPrincipalFromSession(r)
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrNotAuthenticatedSession):
				httpError(w, "no session", http.StatusForbidden, err)
			case errors.Is(err, auth.ErrNoUserSession):
				httpError(w, "no user session", http.StatusForbidden, err)
			default:
				httpError(w, "internal error", http.StatusInternalServerError, err)
			}
			_ = telemetry.RecordError(span, err)
			return
		}
		userId, err := user.GetUuid()
		if err != nil {
			_ = telemetry.RecordError(span, err)
			httpError(w, "error user", http.StatusBadRequest, err)
			return
		}

		liveStream, _, err := getLiveStream(r, streamService)
		if err != nil {
			_ = telemetry.RecordError(span, err)
			handleResourceError(w, err)
			return
		}

		offer, err := getSdpPayload(w, r, webrtc.SDPTypeOffer)
		if err != nil {
			_ = telemetry.RecordError(span, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// track request meta by otel
		span.SetAttributes(
			attribute.String("streamId", liveStream.UUID.String()),
			attribute.String("userId", userId.String()),
		)

		answer, resourceId, err := liveService.CreateLobbyEgressEndpoint(ctx, offer, liveStream, userId)
		if err != nil && errors.Is(err, lobby.ErrSessionAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			_ = telemetry.RecordError(span, err)
			httpError(w, "session already exists", http.StatusConflict, err)
			return
		}

		if err != nil {
			_ = telemetry.RecordError(span, err)
			httpError(w, "error build whep", http.StatusInternalServerError, err)
			return
		}
		span.SetAttributes(attribute.String("sessionId", resourceId))

		response := []byte(answer.SDP)
		hash := md5.Sum(response)

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("etag", fmt.Sprintf("%x", hash))
		w.Header().Set("Location", "resource/"+resourceId)
		contentLen, err := w.Write(response)
		if err != nil {
			_ = telemetry.RecordError(span, err)
			httpError(w, "error build response", http.StatusInternalServerError, err)
			return
		}
		w.Header().Set("Content-Length", strconv.Itoa(contentLen))
	}
}

// old api -----
func whepOffer(streamService *stream.LiveStreamService, liveService *stream.LiveLobbyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := otel.Tracer(tracerName).Start(r.Context(), "whep_offer")
		defer span.End()

		w.Header().Set("Content-Type", "application/sdp")
		user, err := auth.GetPrincipalFromSession(r)
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrNotAuthenticatedSession):
				httpError(w, "no session", http.StatusForbidden, err)
			case errors.Is(err, auth.ErrNoUserSession):
				httpError(w, "no user session", http.StatusForbidden, err)
			default:
				httpError(w, "internal error", http.StatusInternalServerError, err)
			}
			telemetry.RecordError(span, err)
			return
		}

		userId, err := user.GetUuid()
		if err != nil {
			telemetry.RecordError(span, err)
			httpError(w, "error user", http.StatusBadRequest, err)
			return
		}

		liveStream, _, err := getLiveStream(r, streamService)
		if err != nil {
			telemetry.RecordError(span, err)
			handleResourceError(w, err)
			return
		}

		answer, err := liveService.InitLobbyEgressEndpoint(ctx, liveStream, userId)
		if err != nil && errors.Is(err, stream.ErrLobbyNotActive) {
			telemetry.RecordError(span, err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if err != nil {
			telemetry.RecordError(span, err)
			httpError(w, "error build whep", http.StatusInternalServerError, err)
			return
		}

		response := []byte(answer.SDP)
		hash := md5.Sum(response)

		w.WriteHeader(http.StatusCreated)
		contentLen, err := w.Write(response)
		if err != nil {
			telemetry.RecordError(span, err)
			httpError(w, "error build response", http.StatusInternalServerError, err)
			return
		}

		w.Header().Set("etag", fmt.Sprintf("%x", hash))
		w.Header().Set("Content-Length", strconv.Itoa(contentLen))
	}
}
