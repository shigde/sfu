package lobby

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/rtp"
	"github.com/shigde/sfu/pkg/message"
	"golang.org/x/exp/slog"
)

var errTimeoutByWaitingForMessenger = errors.New("timeout by waiting for messenger")
var errSessionClosedByWaitingForMessenger = errors.New("session closed by waiting for messenger")
var waitingTimeOut = 10 * time.Second

type signal struct {
	id                uuid.UUID
	sessionCtx        context.Context
	session           uuid.UUID
	user              uuid.UUID
	egress            *rtp.Endpoint
	ingress           *rtp.Endpoint
	onMuteCbk         func(_ *message.Mute)
	messenger         *messenger
	offerNumber       atomic.Uint32
	receivedMessenger chan struct{}
}

func newSignal(sessionCtx context.Context, session uuid.UUID, user uuid.UUID) *signal {
	return &signal{
		id:                uuid.New(),
		sessionCtx:        sessionCtx,
		session:           session,
		user:              user,
		receivedMessenger: make(chan struct{}),
	}
}

func (s *signal) OnIngressChannel(ingressDC *webrtc.DataChannel) {
	slog.Debug("lobby.signal: get ingress datachannel sender and create messenger", "sessionId", s.session, "userId", s.user)
	s.messenger = newMessenger(ingressDC)
	s.stopWaitingForMessenger()
}

func (s *signal) OnEmptyChannel(_ *webrtc.DataChannel) {
	// we crete an egress data channel because we do not want munging the sdp in case of not added tracks to egress egress
}

func (s *signal) OnHostPipeChannel(egressDC *webrtc.DataChannel) {
	slog.Debug("lobby.signal: get host egress datachannel sender and create messenger", "sessionId", s.session, "userId", s.user)
	s.messenger = newMessenger(egressDC)
	s.stopWaitingForMessenger()
}

func (s *signal) stopWaitingForMessenger() {
	select {
	case <-s.receivedMessenger:
	default:
		close(s.receivedMessenger)
		<-s.receivedMessenger
	}
}

func (s *signal) addEgressEndpoint(endpoint *rtp.Endpoint) {
	s.egress = endpoint
	s.offerNumber.Store(0)
	if s.ingress == nil {
		s.messenger.register(s)
	}
}
func (s *signal) addIngressEndpoint(endpoint *rtp.Endpoint) {
	s.ingress = endpoint
	if s.egress == nil {
		s.messenger.register(s)
	}
}

func (s *signal) OnNegotiationNeeded(offer webrtc.SessionDescription) {
	if _, err := s.messenger.sendOffer(&offer, s.nextOffer()); err != nil {
		slog.Error("lobby.sessionEgressHandler: on negotiated was trigger with error", "err", err, "sessionId", s.session, "user", s.user)
	}
}

func (s *signal) onAnswer(sdp *webrtc.SessionDescription, number uint32) {
	// ignore if offer outdated
	current := s.currentOffer()
	if current != number {
		slog.Debug("lobby.signal: onAnswer ignore", "number", number, "currentNumber", current, "sessionId", s.session, "userId", s.user)
		return
	}

	slog.Debug("lobby.signal: onAnswer set", "number", number, "currentNumber", current, "sessionId", s.session, "user", s.user)

	if err := s.egress.SetAnswer(sdp); err != nil {
		slog.Error("lobby.signal: on answer was trigger with error", "err", err, "sessionId", s.session, "userId", s.user)
	}
	s.egress.SetInitComplete()
}

func (s *signal) onOffer(sdp *webrtc.SessionDescription, responseId uint32, number uint32) {
	slog.Debug("lobby.signal: onAnswer set", "number", number, "sessionId", s.session, "user", s.user)

	answer, err := s.ingress.SetNewOffer(sdp)
	if err != nil {
		slog.Error("lobby.signal: on answer was trigger with error", "err", err, "sessionId", s.session, "userId", s.user)
	}
	if _, err := s.messenger.sendAnswer(answer, responseId, number); err != nil {
		slog.Error("lobby.signal: on answer was trigger with error", "err", err, "sessionId", s.session, "userId", s.user)
	}
}

func (s *signal) onMute(mute *message.Mute) {
	if s.onMuteCbk != nil {
		s.onMuteCbk(mute)
	}
}

func (s *signal) nextOffer() uint32 {
	return s.offerNumber.Add(1)
}

func (s *signal) currentOffer() uint32 {
	return s.offerNumber.Load()
}

func (s *signal) getId() uuid.UUID {
	return s.id
}

func (s *signal) waitForMessengerSetupFinished() <-chan error {
	err := make(chan error)
	go func() {
		defer close(err)
		select {
		case <-s.receivedMessenger:
		case <-s.sessionCtx.Done():
			err <- errSessionClosedByWaitingForMessenger
		case <-time.After(waitingTimeOut):
			err <- errTimeoutByWaitingForMessenger
		}
	}()
	return err
}
