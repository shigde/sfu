package routes

import (
	"crypto/md5"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/auth/request"
	"github.com/shigde/sfu/internal/auth/session"
	"github.com/shigde/sfu/internal/lobby"
	"github.com/shigde/sfu/internal/rest"
	"github.com/shigde/sfu/internal/stream"
	"github.com/shigde/sfu/internal/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func whip(streamService *stream.LiveStreamService, liveService *stream.LiveLobbyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := otel.Tracer(tracerName).Start(r.Context(), "api: whip_create")
		defer span.End()

		w.Header().Set("Content-Type", "application/sdp")

		if err := session.StartSession(w, r); err != nil {
			_ = telemetry.RecordError(span, err)
			rest.HttpError(w, "error", http.StatusInternalServerError, err)
		}

		liveStream, _, err := getLiveStream(r, streamService)
		if err != nil {
			_ = telemetry.RecordError(span, err)
			handleResourceError(w, err)
			return
		}

		offer, err := rest.GetSdpPayload(w, r, webrtc.SDPTypeOffer)
		if err != nil {
			_ = telemetry.RecordError(span, err)
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		user, ok := session.PrincipalFromContext(r.Context())
		if !ok {
			_ = telemetry.RecordError(span, errors.New("no user"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		userId, err := user.GetUuid()
		if err != nil {
			_ = telemetry.RecordError(span, err)
			rest.HttpError(w, "error user", http.StatusBadRequest, err)
			return
		}
		// track request meta by otel
		span.SetAttributes(
			attribute.String("streamId", liveStream.UUID.String()),
			attribute.String("userId", userId.String()),
		)
		request.SetNewRequestToken(w, user.UUID)

		answer, resourceId, err := liveService.CreateLobbyIngressEndpoint(ctx, offer, liveStream, userId)
		if err != nil && errors.Is(err, lobby.ErrSessionAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			_ = telemetry.RecordError(span, err)
			rest.HttpError(w, "session already exists", http.StatusConflict, err)
			return
		}

		span.SetAttributes(attribute.String("sessionId", resourceId))

		if err != nil {
			_ = telemetry.RecordError(span, err)
			rest.HttpError(w, "error build whip", http.StatusInternalServerError, err)
			return
		}

		response := []byte(answer.SDP)
		hash := md5.Sum(response)

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("etag", fmt.Sprintf("%x", hash))
		w.Header().Set("Location", "resource/"+resourceId)
		contentLen, err := w.Write(response)
		if err != nil {
			_ = telemetry.RecordError(span, err)
			rest.HttpError(w, "error build response", http.StatusInternalServerError, err)
			return
		}
		w.Header().Set("Content-Length", strconv.Itoa(contentLen))
	}
}

func whipDelete(streamService *stream.LiveStreamService, liveService *stream.LiveLobbyService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := otel.Tracer(tracerName).Start(r.Context(), "api: whip_delete")
		defer span.End()

		w.Header().Set("Content-Type", "application/sdp")
		user, err := session.GetPrincipalFromSession(r)
		if err != nil {
			switch {
			case errors.Is(err, session.ErrNotAuthenticatedSession):
				rest.HttpError(w, "no session", http.StatusForbidden, err)
			case errors.Is(err, session.ErrNoUserSession):
				rest.HttpError(w, "no user session", http.StatusForbidden, err)
			default:
				rest.HttpError(w, "internal error", http.StatusInternalServerError, err)
			}
			_ = telemetry.RecordError(span, err)
			return
		}

		userId, err := user.GetUuid()
		if err != nil {
			_ = telemetry.RecordError(span, err)
			rest.HttpError(w, "error user", http.StatusBadRequest, err)
			return
		}

		liveStream, _, err := getLiveStream(r, streamService)
		if err != nil {
			_ = telemetry.RecordError(span, err)
			handleResourceError(w, err)
			return
		}

		left, err := liveService.LeaveLobby(ctx, liveStream, userId)
		if err != nil {
			_ = telemetry.RecordError(span, err)
			if errors.Is(err, stream.ErrLobbyNotActive) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			rest.HttpError(w, "error", http.StatusInternalServerError, err)
			return
		}

		if !left {
			rest.HttpError(w, "error", http.StatusInternalServerError, errors.New("leaving lobby was not possible"))
			return
		}

		if err := session.DeleteSession(w, r); err != nil {
			_ = telemetry.RecordError(span, err)
			rest.HttpError(w, "error", http.StatusInternalServerError, err)
		}

		w.WriteHeader(http.StatusOK)
	}
}