package lobby

import (
	"fmt"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/rtp"
	"golang.org/x/exp/slog"
)

type senderHandler struct {
	id          uuid.UUID
	session     uuid.UUID
	user        uuid.UUID
	endpoint    *rtp.Endpoint
	messenger   *messenger
	offerNumber atomic.Uint32
}

func newSenderHandler(session uuid.UUID, user uuid.UUID, messenger *messenger) *senderHandler {
	h := &senderHandler{
		id:        uuid.New(),
		session:   session,
		user:      user,
		messenger: messenger,
	}
	h.offerNumber.Store(0)
	messenger.register(h)
	return h

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
	if _, err := h.messenger.sendOffer(&offer, h.nextOffer()); err != nil {
		slog.Error("lobby.senderHandler: on negotiated was trigger with error", "err", err, "session", h.session, "user", h.user)
	}
}

func (h *senderHandler) onAnswer(sdp *webrtc.SessionDescription, number uint32) {
	// ignore if offer outdated
	current := h.currentOffer()
	if current != number {
		slog.Debug("lobby.senderHandler: onAnswer ignore", "number", number, "currentNumber", current, "session", h.session, "user", h.user)
		return
	}

	slog.Debug("lobby.senderHandler: onAnswer set", "number", number, "currentNumber", current, "session", h.session, "user", h.user)

	if err := h.endpoint.SetAnswer(sdp); err != nil {
		slog.Error("lobby.senderHandler: on answer was trigger with error", "err", err, "session", h.session, "user", h.user)
	}
}

func (h *senderHandler) OnChannel(_ *webrtc.DataChannel) {
	slog.Debug("lobby.senderHandler: datachannel is open but we do not need them", "session", h.session, "user", h.user)

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

func (h *senderHandler) nextOffer() uint32 {
	return h.offerNumber.Add(1)
}

func (h *senderHandler) currentOffer() uint32 {
	return h.offerNumber.Load()
}

func (h *senderHandler) getId() uuid.UUID {
	return h.id
}
