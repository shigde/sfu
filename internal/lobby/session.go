package lobby

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/slog"
)

var (
	errSessionAlreadyClosed           = errors.New("the sessions was already closed")
	errSessionCouldNotClosed          = errors.New("the sessions could not closed in right way")
	errReceiverInSessionAlreadyExists = errors.New("receiver already exists")
	errNoReceiverInSession            = errors.New("no receiver in session")
	errSenderInSessionAlreadyExists   = errors.New("sender already exists")
	errNoSenderInSession              = errors.New("no sender exists")
	errSessionRequestTimeout          = errors.New("session request timeout error")
)

var sessionReqTimeout = 3 * time.Second

type session struct {
	Id               uuid.UUID
	user             uuid.UUID
	rtpEngine        rtpEngine
	hub              *hub
	receiver         *receiverHandler
	sender           *senderHandler
	reqChan          chan *sessionRequest
	quit             chan struct{}
	onInternallyQuit onInternallyQuit
}

func newSession(user uuid.UUID, hub *hub, engine rtpEngine, onQuit onInternallyQuit) *session {
	quit := make(chan struct{})
	requests := make(chan *sessionRequest)

	session := &session{
		Id:               uuid.New(),
		user:             user,
		rtpEngine:        engine,
		hub:              hub,
		reqChan:          requests,
		quit:             quit,
		onInternallyQuit: onQuit,
	}

	go session.run()
	return session
}

func (s *session) run() {
	slog.Info("lobby.sessions: run", "id", s.Id, "user", s.user)
	for {
		select {
		case req := <-s.reqChan:
			s.handleSessionReq(req)
		case <-s.quit:
			// @TODO Take care that's every stream is closed!
			slog.Info("lobby.sessions: stop running", "id", s.Id, "user", s.user)
			return
		}
	}
}

func (s *session) runRequest(req *sessionRequest) {
	slog.Debug("lobby.sessions: runRequest", "id", s.Id, "user", s.user)
	select {
	case s.reqChan <- req:
		slog.Debug("lobby.sessions: runRequest - return response", "id", s.Id, "user", s.user)
	case <-s.quit:
		req.err <- errSessionAlreadyClosed
		slog.Debug("lobby.sessions: runRequest - interrupted because sessions closed", "id", s.Id, "user", s.user)
	case <-time.After(sessionReqTimeout):
		req.err <- errSessionRequestTimeout
		slog.Error("lobby.sessions: runRequest - interrupted because request timeout", "id", s.Id, "user", s.user)
	}
}

func (s *session) handleSessionReq(req *sessionRequest) {
	slog.Info("lobby.sessions: handle session req", "id", s.Id, "user", s.user)

	var sdp *webrtc.SessionDescription
	var err error
	switch req.sessionReqType {
	case offerReq:
		sdp, err = s.handleOfferReq(req)
	case answerReq:
		sdp, err = s.handleAnswerReq(req)
	case startReq:
		sdp, err = s.handleStartReq(req)
	}
	if err != nil {
		req.err <- fmt.Errorf("handle request: %w", err)
		return
	}

	req.respSDPChan <- sdp
}

func (s *session) handleOfferReq(req *sessionRequest) (*webrtc.SessionDescription, error) {
	ctx, span := otel.Tracer(tracerName).Start(req.ctx, "session:handleOfferReq")
	defer span.End()

	if s.receiver != nil {
		return nil, errReceiverInSessionAlreadyExists
	}
	s.receiver = newReceiverHandler(s.Id, s.user, s.onInternallyQuit)
	endpoint, err := s.rtpEngine.NewReceiverEndpoint(ctx, *req.reqSDP, s.hub, s.receiver)
	if err != nil {
		return nil, fmt.Errorf("create rtp connection: %w", err)
	}
	s.receiver.endpoint = endpoint
	answer, err := s.receiver.endpoint.GetLocalDescription(ctx)
	if err != nil {
		return nil, fmt.Errorf("create rtp answerReq: %w", err)
	}
	return answer, nil
}

func (s *session) handleAnswerReq(req *sessionRequest) (*webrtc.SessionDescription, error) {
	_, span := otel.Tracer(tracerName).Start(req.ctx, "session:handleAnswerReq")
	defer span.End()

	if s.sender == nil || s.sender.endpoint == nil {
		return nil, errNoSenderInSession
	}

	s.sender.onAnswer(req.reqSDP, 0)
	return nil, nil
}

func (s *session) handleStartReq(req *sessionRequest) (*webrtc.SessionDescription, error) {
	ctx, span := otel.Tracer(tracerName).Start(req.ctx, "session:handleStartReq")
	defer span.End()

	if s.sender != nil {
		return nil, errSenderInSessionAlreadyExists
	}

	if s.receiver == nil {
		return nil, errNoReceiverInSession
	}

	select {
	case err := <-s.receiver.waitForMessenger():
		if err != nil {
			return nil, fmt.Errorf("waiting for messenger: %w", err)
		}
	case <-s.quit:
		return nil, errSessionAlreadyClosed
	case <-time.After(sessionReqTimeout):
		return nil, errSessionRequestTimeout
	}

	s.sender = newSenderHandler(s.Id, s.user, s.receiver.messenger)

	trackList, err := s.hub.getTrackList()
	if err != nil {
		return nil, fmt.Errorf("reading track list by creating rtp connection: %w", err)
	}

	endpoint, err := s.rtpEngine.NewSenderEndpoint(ctx, trackList, s.sender)
	if err != nil {
		return nil, fmt.Errorf("create rtp connection: %w", err)
	}
	s.sender.endpoint = endpoint

	offer, err := s.sender.endpoint.GetLocalDescription(ctx)
	if err != nil {
		return nil, fmt.Errorf("create rtp answerReq: %w", err)
	}

	return offer, nil
}

func (s *session) stop() error {
	slog.Info("lobby.sessions: stop", "id", s.Id, "user", s.user)
	select {
	case <-s.quit:
		slog.Error("lobby.sessions: the rtp sessions was already closed", "sessionId", s.Id, "user", s.user)
		return errSessionAlreadyClosed
	default:
		close(s.quit)
		slog.Info("lobby.sessions: stopped was triggered", "sessionId", s.Id, "user", s.user)
		<-s.quit
		if s.sender != nil {
			if err := s.sender.close(); err != nil {
				slog.Error("lobby.sessions: closing sender", "sessionId", s.Id, "user", s.user)
			}
		}

		if s.receiver != nil {
			if err := s.receiver.close(); err != nil {
				slog.Error("lobby.sessions: closing receiver", "sessionId", s.Id, "user", s.user)
			}
		}
	}
	return nil
}

func (s *session) addTrack(track *webrtc.TrackLocalStaticRTP) {
	if s.sender != nil && s.sender.endpoint != nil {
		s.sender.endpoint.AddTrack(track)
	}
}

func (s *session) removeTrack(track *webrtc.TrackLocalStaticRTP) {
	if s.sender != nil && s.sender.endpoint != nil {
		s.sender.endpoint.RemoveTrack(track)
	}
}

type onInternallyQuit = func(ctx context.Context, user uuid.UUID) bool
