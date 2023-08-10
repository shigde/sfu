package media

import (
	"crypto/md5"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/auth"
	"github.com/shigde/sfu/internal/telemetry"
	"go.opentelemetry.io/otel"
)

func whip(spaceManager spaceGetCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := otel.Tracer(tracerName).Start(r.Context(), "whip-create")
		//ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer(tracerName).Start(ctx, "post-whip")
		defer span.End()

		w.Header().Set("Content-Type", "application/sdp")

		if err := auth.StartSession(w, r); err != nil {
			telemetry.RecordError(span, err)
			httpError(w, "error", http.StatusInternalServerError, err)
		}

		liveStream, space, err := getLiveStream(r, spaceManager)
		if err != nil {
			telemetry.RecordError(span, err)
			handleResourceError(w, err)
			return
		}

		offer, err := getSdpPayload(w, r, webrtc.SDPTypeOffer)
		if err != nil {
			telemetry.RecordError(span, err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		user, ok := auth.PrincipalFromContext(r.Context())
		if !ok {
			telemetry.RecordError(span, errors.New("noe user"))
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

		answer, resourceId, err := space.EnterLobby(ctx, offer, liveStream, userId)
		if err != nil {
			telemetry.RecordError(span, err)
			httpError(w, "error build whip", http.StatusInternalServerError, err)
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
