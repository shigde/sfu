package lobby

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/rtp"
	"golang.org/x/exp/slog"
)

type receiverHandler struct {
	session         uuid.UUID
	user            uuid.UUID
	endpoint        *rtp.Endpoint
	messenger       *messenger
	onEndpointClose onInternallyQuit
}

func newReceiverHandler(session uuid.UUID, user uuid.UUID, onEndpointClose onInternallyQuit) *receiverHandler {
	return &receiverHandler{
		session:         session,
		user:            user,
		onEndpointClose: onEndpointClose,
	}
}

func (h *receiverHandler) OnConnectionStateChange(state webrtc.ICEConnectionState) {
	if state == webrtc.ICEConnectionStateFailed {
		slog.Warn("lobby.receiverHandler: endpoint become idle", "session", h.session, "user", h.user)
	}

	if state == webrtc.ICEConnectionStateDisconnected || state == webrtc.ICEConnectionStateClosed {
		slog.Warn("lobby.receiverHandler: endpoint lost connection", "session", h.session, "user", h.user)
		h.onEndpointClose(context.Background(), h.user)
	}
}

func (h *receiverHandler) OnNegotiationNeeded(offer webrtc.SessionDescription) {
	slog.Warn("lobby.receiverHandler: on negotiated was trigger for static connection", "session", h.session, "user", h.user)
}
func (h *receiverHandler) OnOnChannel(dc *webrtc.DataChannel) {
	h.messenger = newMessenger(dc)
}

func (h *receiverHandler) close() error {
	if h.endpoint == nil {
		return nil
	}

	if err := h.endpoint.Close(); err != nil {
		return fmt.Errorf("receiver handler closing endpoint: %w", err)
	}
	return nil
}