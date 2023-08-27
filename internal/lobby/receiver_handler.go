package lobby

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/rtp"
	"golang.org/x/exp/slog"
)

var errTimeoutByWaitingForMessenger = errors.New("timeout by waiting for messenger")
var waitingTimeOut = 3 * time.Second

type receiverHandler struct {
	session           uuid.UUID
	user              uuid.UUID
	endpoint          *rtp.Endpoint
	messenger         *messenger
	onEndpointClose   onInternallyQuit
	receivedMessenger chan struct{}
}

func newReceiverHandler(session uuid.UUID, user uuid.UUID, onEndpointClose onInternallyQuit) *receiverHandler {
	return &receiverHandler{
		session:           session,
		user:              user,
		onEndpointClose:   onEndpointClose,
		receivedMessenger: make(chan struct{}),
	}
}

func (h *receiverHandler) OnConnectionStateChange(state webrtc.ICEConnectionState) {
	if state == webrtc.ICEConnectionStateFailed {
		slog.Warn("lobby.receiverHandler: endpoint become idle", "session", h.session, "user", h.user)
	}

	if state == webrtc.ICEConnectionStateClosed {
		slog.Info("lobby.receiverHandler: connection closed", "session", h.session, "user", h.user)
	}

	if state == webrtc.ICEConnectionStateDisconnected {
		slog.Warn("lobby.receiverHandler: endpoint lost connection", "session", h.session, "user", h.user)
		h.onEndpointClose(context.Background(), h.user)
	}
}

func (h *receiverHandler) OnNegotiationNeeded(offer webrtc.SessionDescription) {
	slog.Warn("lobby.receiverHandler: on negotiated was trigger for static connection", "session", h.session, "user", h.user)
}
func (h *receiverHandler) OnChannel(dc *webrtc.DataChannel) {
	slog.Debug("lobby.receiveHandler: get an datachannel sender and create messenger", "session", h.session, "user", h.user)
	h.messenger = newMessenger(dc)
	h.stopWaitingForMessenger()
}

func (h *receiverHandler) stopWaitingForMessenger() {
	select {
	case <-h.receivedMessenger:
	default:
		close(h.receivedMessenger)
		<-h.receivedMessenger
	}
}

func (h *receiverHandler) close() error {
	h.stopWaitingForMessenger()

	if h.endpoint != nil {
		if err := h.endpoint.Close(); err != nil {
			return fmt.Errorf("receiver handler closing endpoint: %w", err)
		}
	}

	if h.messenger != nil {
		h.messenger.close()
	}

	return nil
}

func (h *receiverHandler) waitForMessenger() <-chan error {
	err := make(chan error)
	go func() {
		defer close(err)
		select {
		case <-h.receivedMessenger:
		case <-time.After(waitingTimeOut):
			err <- errTimeoutByWaitingForMessenger
		}
	}()
	return err
}
