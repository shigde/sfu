package sessions

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/lobby/clients"
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
	offerer           *rtp.Endpoint // The offerer is always an egress endpoint or nil
	answerer          *rtp.Endpoint // The answerer is always an ingress endpoint or nil
	onMuteCbk         func(_ *message.Mute)
	messenger         *clients.Messenger
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

func (s *signal) OnSenderChannel(channel *webrtc.DataChannel) {
	slog.Debug("signal: get signal channel and create messenger", "sessionId", s.session, "userId", s.user)
	s.messenger = clients.NewMessenger(channel)
	// we register this signaler for datachannel messages after we received a webrtc channel
	s.messenger.Register(s)
	s.stopWaitingForMessenger()
}

func (s *signal) OnSilentChannel(_ *webrtc.DataChannel) {
	// we crete an egress data channel because we do not want munging the sdp in case of not added tracks to egress egress
}

func (s *signal) stopWaitingForMessenger() {
	select {
	case <-s.receivedMessenger:
	default:
		close(s.receivedMessenger)
		<-s.receivedMessenger
	}
}

func (s *signal) setAnswerer(endpoint *rtp.Endpoint) {
	s.answerer = endpoint
}

func (s *signal) setOfferer(endpoint *rtp.Endpoint) {
	s.offerer = endpoint
	s.offerNumber.Store(0)
}

func (s *signal) OnNegotiationNeeded(offer webrtc.SessionDescription) {
	if _, err := s.messenger.SendOffer(&offer, s.nextOffer()); err != nil {
		slog.Error("lobby.sessionEgressHandler: on negotiated was trigger with error", "err", err, "sessionId", s.session, "user", s.user)
	}
}

func (s *signal) OnAnswer(sdp *webrtc.SessionDescription, number uint32) {
	if s.offerer == nil {
		slog.Warn("lobby.signal: no offerer exists to get this answer onAnswer", "number", number, "sessionId", s.session, "user", s.user)
		return
	}

	// ignore if offer outdated
	current := s.currentOffer()
	if current != number {
		slog.Debug("lobby.signal: onAnswer ignore", "number", number, "currentNumber", current, "sessionId", s.session, "userId", s.user)
		return
	}
	slog.Debug("lobby.signal: onAnswer set", "number", number, "currentNumber", current, "sessionId", s.session, "user", s.user)

	if err := s.offerer.SetAnswer(sdp); err != nil {
		slog.Error("lobby.signal: on answer was trigger with error", "err", err, "sessionId", s.session, "userId", s.user)
	}
	s.offerer.SetInitComplete()
}

func (s *signal) OnOffer(sdp *webrtc.SessionDescription, responseId uint32, number uint32) {
	slog.Debug("lobby.signal: onAnswer set", "number", number, "sessionId", s.session, "user", s.user)

	if s.answerer == nil {
		slog.Warn("lobby.signal: no answerer exists to answer this offer onOffer", "number", number, "sessionId", s.session, "user", s.user)
		return
	}

	answer, err := s.answerer.SetNewOffer(sdp)
	if err != nil {
		slog.Error("lobby.signal: on answer was trigger with error", "err", err, "sessionId", s.session, "userId", s.user)
	}
	if _, err := s.messenger.SendAnswer(answer, responseId, number); err != nil {
		slog.Error("lobby.signal: on answer was trigger with error", "err", err, "sessionId", s.session, "userId", s.user)
	}
}

func (s *signal) OnMute(mute *message.Mute) {
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

func (s *signal) GetId() uuid.UUID {
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
