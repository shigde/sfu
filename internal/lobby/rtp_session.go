package lobby

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"golang.org/x/exp/slog"
)

var errRtpSessionAlreadyClosed = errors.New("the rtp session was already closed")
var errOfferInterrupted = errors.New("request an offer get interrupted")

var sessionReqTimeout = 3 * time.Second

type rtpSession struct {
	Id        uuid.UUID
	user      uuid.UUID
	streams   *rtpStreamRepository
	rtpEngine rtpEngine
	offerChan chan *offerRequest
	quit      chan struct{}
}

func newRtpSession(user uuid.UUID, e rtpEngine) *rtpSession {
	repo := newRtpStreamRepository()
	q := make(chan struct{})
	o := make(chan *offerRequest)
	session := &rtpSession{
		Id:        uuid.New(),
		user:      user,
		streams:   repo,
		rtpEngine: e,
		offerChan: o,
		quit:      q,
	}

	go session.run()
	return session
}

func (s *rtpSession) run() {
	slog.Info("lobby.rtpSession: run", "id", s.Id, "user", s.user)
	for {
		select {
		case offer := <-s.offerChan:
			s.handleOffer(offer)
		case <-s.quit:
			// @TODO Take care that's every stream is closed!
			slog.Info("lobby.rtpSession: stop running", "id", s.Id, "user", s.user)
			return
		}
	}
}

func (s *rtpSession) runOffer(offerReq *offerRequest) {
	slog.Debug("lobby.rtpSession: offer", "id", s.Id, "user", s.user)
	select {
	case s.offerChan <- offerReq:
		slog.Debug("lobby.rtpSession: offer - offer requested", "id", s.Id, "user", s.user)
	case <-s.quit:
		offerReq.err <- errRtpSessionAlreadyClosed
		slog.Debug("lobby.rtpSession: offer - interrupted because session closed", "id", s.Id, "user", s.user)
	case <-time.After(sessionReqTimeout):
		slog.Error("lobby.rtpSession: offer - interrupted because request timeout", "id", s.Id, "user", s.user)
	}
}

func (s *rtpSession) handleOffer(offerReq *offerRequest) {
	slog.Info("lobby.rtpSession: handle offer", "id", s.Id, "user", s.user)
	stream := newRtpStream()
	s.streams.Add(stream)
	conn, err := s.rtpEngine.NewConnection(*offerReq.offer, stream.Id)
	if err != nil {
		offerReq.err <- fmt.Errorf("create rtp connection: %w", err)
		return
	}

	answer, err := conn.GetAnswer(offerReq.ctx)
	if err != nil {
		offerReq.err <- fmt.Errorf("create rtp answer: %w", err)
		return
	}
	offerReq.answer <- answer
}

func (s *rtpSession) stop() error {
	slog.Info("lobby.rtpSession: stop", "id", s.Id, "user", s.user)
	select {
	case <-s.quit:
		slog.Error("lobby.rtpSession: the rtp session was already closed", "id", s.Id, "user", s.user)
		return errRtpSessionAlreadyClosed
	default:
		close(s.quit)
		slog.Info("lobby.rtpSession: stopped was triggered", "id", s.Id, "user", s.user)
	}
	return nil
}

type offerRequest struct {
	offer  *webrtc.SessionDescription
	answer chan *webrtc.SessionDescription
	err    chan error
	ctx    context.Context
}

func newOfferRequest(ctx context.Context, offer *webrtc.SessionDescription) *offerRequest {
	return &offerRequest{
		offer:  offer,
		answer: make(chan *webrtc.SessionDescription),
		err:    make(chan error),
		ctx:    ctx,
	}
}
