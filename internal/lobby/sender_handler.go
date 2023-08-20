package lobby

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/rtp"
	"golang.org/x/exp/slog"
)

type senderHandler struct {
	session   uuid.UUID
	user      uuid.UUID
	endpoint  *rtp.Endpoint
	messenger *messenger
}

func newSenderHandler(session uuid.UUID, user uuid.UUID, messenger *messenger) *senderHandler {
	return &senderHandler{
		session:   session,
		user:      user,
		messenger: messenger,
	}
}

func (h *senderHandler) OnConnectionStateChange(state webrtc.ICEConnectionState) {
	if state == webrtc.ICEConnectionStateFailed {
		slog.Warn("lobby.senderHandler: endpoint become idle", "session", h.session, "user", h.user)
	}

	if state == webrtc.ICEConnectionStateDisconnected || state == webrtc.ICEConnectionStateClosed {
		slog.Warn("lobby.senderHandler: endpoint lost connection", "session", h.session, "user", h.user)
	}
}

func (h *senderHandler) OnNegotiationNeeded(offer webrtc.SessionDescription) {
	if _, err := h.messenger.sendOffer(&offer, 1); err != nil {
		slog.Error("lobby.senderHandler: on negotiated was trigger with error", "err", err, "session", h.session, "user", h.user)
	}
}
func (h *senderHandler) OnOnChannel(dc *webrtc.DataChannel) {
	slog.Debug("lobby.senderHandler: datachannel is open", "session", h.session, "user", h.user)

}

func (h *senderHandler) close() error {
	if h.endpoint == nil {
		return nil
	}

	if err := h.endpoint.Close(); err != nil {
		return fmt.Errorf("sender handler closing endpoint: %w", err)
	}
	return nil
}
