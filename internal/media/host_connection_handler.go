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
)

func openPipe(streamService *stream.LiveStreamService, liveService *stream.LiveLobbyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := otel.Tracer(tracerName).Start(r.Context(), "open-host-pipe")
		defer span.End()

		w.Header().Set("Content-Type", "application/sdp")

		liveStream, _, err := getLiveStream(r, streamService)
		if err != nil {
			telemetry.RecordError(span, err)
			handleResourceError(w, err)
			return
		}

		offer, err := getSdpPayload(w, r, webrtc.SDPTypeOffer)
		if err != nil {
			telemetry.RecordError(span, err)
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		user, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			telemetry.RecordError(span, errors.New("no user"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		userId, err := user.GetUuid()
		if err != nil {
			telemetry.RecordError(span, err)
			httpError(w, "error user", http.StatusBadRequest, err)
			return
		}
		auth.SetNewRequestToken(w, user.UUID)

		answer, resourceId, err := liveService.CreateLobbyHostConnection(ctx, offer, liveStream, userId)
		if err != nil && errors.Is(err, lobby.ErrSessionAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			telemetry.RecordError(span, err)
			httpError(w, "session already exists", http.StatusConflict, err)
			return
		}

		if err != nil {
			telemetry.RecordError(span, err)
			httpError(w, "error build host pipe", http.StatusInternalServerError, err)
			return
		}

		response := []byte(answer.SDP)
		hash := md5.Sum(response)

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("etag", fmt.Sprintf("%x", hash))
		w.Header().Set("Location", "resource/"+resourceId)
		contentLen, err := w.Write(response)
		if err != nil {
			telemetry.RecordError(span, err)
			httpError(w, "error build response", http.StatusInternalServerError, err)
			return
		}
		w.Header().Set("Content-Length", strconv.Itoa(contentLen))
	}
}

func openHostIngress(streamService *stream.LiveStreamService, liveService *stream.LiveLobbyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := otel.Tracer(tracerName).Start(r.Context(), "open-host-Ingress")
		defer span.End()

		w.Header().Set("Content-Type", "application/sdp")

		liveStream, _, err := getLiveStream(r, streamService)
		if err != nil {
			telemetry.RecordError(span, err)
			handleResourceError(w, err)
			return
		}

		offer, err := getSdpPayload(w, r, webrtc.SDPTypeOffer)
		if err != nil {
			telemetry.RecordError(span, err)
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		user, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			telemetry.RecordError(span, errors.New("no user"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		userId, err := user.GetUuid()
		if err != nil {
			telemetry.RecordError(span, err)
			httpError(w, "error user", http.StatusBadRequest, err)
			return
		}
		auth.SetNewRequestToken(w, user.UUID)

		answer, resourceId, err := liveService.CreateLobbyHostIngressConnection(ctx, offer, liveStream, userId)
		if err != nil && errors.Is(err, lobby.ErrSessionAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			telemetry.RecordError(span, err)
			httpError(w, "session already exists", http.StatusConflict, err)
			return
		}

		if err != nil {
			telemetry.RecordError(span, err)
			httpError(w, "error build host pipe", http.StatusInternalServerError, err)
			return
		}

		response := []byte(answer.SDP)
		hash := md5.Sum(response)

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("etag", fmt.Sprintf("%x", hash))
		w.Header().Set("Location", "resource/"+resourceId)
		contentLen, err := w.Write(response)
		if err != nil {
			telemetry.RecordError(span, err)
			httpError(w, "error build response", http.StatusInternalServerError, err)
			return
		}
		w.Header().Set("Content-Length", strconv.Itoa(contentLen))
	}
}

func closePipe(streamService *stream.LiveStreamService, liveService *stream.LiveLobbyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := otel.Tracer(tracerName).Start(r.Context(), "close host pipe")
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

		left, err := liveService.CloseLobbyHostConnection(ctx, liveStream, userId)
		if err != nil {
			telemetry.RecordError(span, err)
			if errors.Is(err, stream.ErrLobbyNotActive) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			httpError(w, "error", http.StatusInternalServerError, err)
			return
		}

		if !left {
			httpError(w, "error", http.StatusInternalServerError, errors.New("close host pipe lobby not possible"))
			return
		}

		if err := auth.DeleteSession(w, r); err != nil {
			telemetry.RecordError(span, err)
			httpError(w, "error", http.StatusInternalServerError, err)
		}

		w.WriteHeader(http.StatusOK)
	}
}
