package lobby

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/rtp"
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
	errUnknownSessionRequestType      = errors.New("unknown session request type")
)

var (
	sessionReqTimeout = 3 * time.Second

	iceGatheringTimeout = 2 * time.Second
)

type session struct {
	Id               uuid.UUID
	user             uuid.UUID
	rtpEngine        rtpEngine
	hub              *hub
	receiver         *receiverHandler
	sender           *senderHandler
	reqChan          chan *sessionRequest
	quit             chan struct{}
	onInternallyQuit chan<- uuid.UUID
}

func newSession(user uuid.UUID, hub *hub, engine rtpEngine, onInternallyQuit chan<- uuid.UUID) *session {
	quit := make(chan struct{})
	requests := make(chan *sessionRequest)

	session := &session{
		Id:               uuid.New(),
		user:             user,
		rtpEngine:        engine,
		hub:              hub,
		reqChan:          requests,
		quit:             quit,
		onInternallyQuit: onInternallyQuit,
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
	case offerIngressReq:
		sdp, err = s.handleOfferIngressReq(req)
	case initEgressReq:
		sdp, err = s.handleInitEgressReq(req)
	case answerEgressReq:
		sdp, err = s.handleAnswerEgressReq(req)
	case offerStaticEgressReq:
		sdp, err = s.handleOfferStaticEgressReq(req)
	default:
		err = errUnknownSessionRequestType
	}
	if err != nil {
		req.err <- fmt.Errorf("handle request: %w", err)
		return
	}

	req.respSDPChan <- sdp
}

func (s *session) handleOfferIngressReq(req *sessionRequest) (*webrtc.SessionDescription, error) {
	ctx, span := otel.Tracer(tracerName).Start(req.ctx, "session:handleOfferIngressReq")
	defer span.End()

	if s.receiver != nil {
		return nil, errReceiverInSessionAlreadyExists
	}
	s.receiver = newReceiverHandler(s.Id, s.user, func(ctx context.Context, user uuid.UUID) bool {
		go func() {
			select {
			case s.onInternallyQuit <- user:
				slog.Debug("lobby.sessions: internally quit of session", "session id", s.Id, "user", s.user)
			case <-s.quit:
				slog.Debug("lobby.sessions: internally quit interrupted because session already closed", "session id", s.Id, "user", s.user)
			}
		}()
		return true
	})

	option := make([]rtp.EndpointOption, 0)
	option = append(option, rtp.EndpointWithDataChannel(s.receiver.OnChannel))
	option = append(option, rtp.EndpointWithConnectionStateListener(s.receiver.OnConnectionStateChange))
	option = append(option, rtp.EndpointWithTrackDispatcher(s.hub))

	endpoint, err := s.rtpEngine.EstablishIngressEndpoint(ctx, s.Id, *req.reqSDP, option...)
	if err != nil {
		return nil, fmt.Errorf("create rtp connection: %w", err)
	}
	s.receiver.endpoint = endpoint
	ctxTimeout, cancel := context.WithTimeout(ctx, iceGatheringTimeout)
	defer cancel()
	answer, err := s.receiver.endpoint.GetLocalDescription(ctxTimeout)
	if err != nil {
		return nil, fmt.Errorf("create rtp answerEgressReq: %w", err)
	}
	return answer, nil
}

func (s *session) handleAnswerEgressReq(req *sessionRequest) (*webrtc.SessionDescription, error) {
	_, span := otel.Tracer(tracerName).Start(req.ctx, "session:handleAnswerEgressReq")
	defer span.End()

	if s.sender == nil || s.sender.endpoint == nil {
		return nil, errNoSenderInSession
	}

	s.sender.onAnswer(req.reqSDP, 0)
	return nil, nil
}

func (s *session) handleInitEgressReq(req *sessionRequest) (*webrtc.SessionDescription, error) {
	ctx, span := otel.Tracer(tracerName).Start(req.ctx, "session:handleInitEgressReq")
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

	trackList, err := s.hub.getTrackList(filterForSession(s.Id))
	if err != nil {
		return nil, fmt.Errorf("reading track list by creating rtp connection: %w", err)
	}
	option := make([]rtp.EndpointOption, 0)
	for _, track := range trackList {
		option = append(option, rtp.EndpointWithTrack(track.GetTrackLocal(), track.GetPurpose()))
	}
	s.sender = newSenderHandler(s.Id, s.user, s.receiver.messenger)
	option = append(option, rtp.EndpointWithDataChannel(s.sender.OnChannel))
	option = append(option, rtp.EndpointWithNegotiationNeededListener(s.sender.OnNegotiationNeeded))
	option = append(option, rtp.EndpointWithConnectionStateListener(s.sender.OnConnectionStateChange))

	endpoint, err := s.rtpEngine.EstablishEgressEndpoint(ctx, s.Id, option...)
	if err != nil {
		return nil, fmt.Errorf("create rtp connection: %w", err)
	}
	s.sender.endpoint = endpoint

	ctxTimeout, cancel := context.WithTimeout(ctx, iceGatheringTimeout)
	defer cancel()
	offer, err := s.sender.endpoint.GetLocalDescription(ctxTimeout)
	if err != nil {
		return nil, fmt.Errorf("create rtp answerEgressReq: %w", err)
	}

	return offer, nil
}

func (s *session) handleOfferStaticEgressReq(req *sessionRequest) (*webrtc.SessionDescription, error) {
	ctx, span := otel.Tracer(tracerName).Start(req.ctx, "session:handleOfferStaticEgressReq")
	defer span.End()

	// This sender Handler has no message handler.
	// Without a message handler this sender is only a placeholder for the endpoint
	s.sender = &senderHandler{
		id:      uuid.New(),
		session: s.Id,
		user:    s.user,
	}

	trackList, err := s.hub.getTrackList(filterForSession(s.Id))
	if err != nil {
		return nil, fmt.Errorf("reading track list by creating rtp connection: %w", err)
	}
	option := make([]rtp.EndpointOption, len(trackList))

	for _, track := range trackList {
		option = append(option, rtp.EndpointWithTrack(track.GetTrackLocal(), track.GetPurpose()))
	}

	endpoint, err := s.rtpEngine.EstablishStaticEgressEndpoint(ctx, s.Id, *req.reqSDP, option...)

	if err != nil {
		return nil, fmt.Errorf("create rtp connection: %w", err)
	}
	s.sender.endpoint = endpoint
	ctxTimeout, cancel := context.WithTimeout(ctx, iceGatheringTimeout)
	defer cancel()
	answer, err := s.sender.endpoint.GetLocalDescription(ctxTimeout)
	if err != nil {
		return nil, fmt.Errorf("create rtp answerEgressReq: %w", err)
	}
	return answer, nil
}

func (s *session) stop() error {
	slog.Info("lobby.sessions: stop", "id", s.Id, "user", s.user)
	select {
	case <-s.quit:
		slog.Error("lobby.sessions: the rtp sessions was already closed", "sessionId", s.Id, "user", s.user)
		return errSessionAlreadyClosed
	default:

		if s.sender != nil {
			if err := s.sender.close(); err != nil {
				slog.Error("lobby.sessions: closing sender", "err", err, "sessionId", s.Id, "user", s.user)
			}
		}

		if s.receiver != nil {
			if err := s.receiver.close(); err != nil {
				slog.Error("lobby.sessions: closing receiver", "err", err, "sessionId", s.Id, "user", s.user)
			}
		}
		close(s.quit)
		slog.Info("lobby.sessions: stopped was triggered", "sessionId", s.Id, "user", s.user)
		<-s.quit
	}
	return nil
}

func (s *session) addTrack(trackInfo *rtp.TrackInfo) {
	track := trackInfo.GetTrackLocal()
	slog.Debug("lobby.sessions: add track", "trackId", track.ID(), "streamId", track.StreamID(), "sessionId", s.Id, "user", s.user)
	if s.sender != nil && s.sender.endpoint != nil {
		slog.Debug("lobby.sessions: add track - to egress endpoint", "trackId", track.ID(), "streamId", track.StreamID(), "sessionId", s.Id, "user", s.user)
		s.sender.endpoint.AddTrack(track, trackInfo.Purpose)
	}
}

func (s *session) removeTrack(trackInfo *rtp.TrackInfo) {
	track := trackInfo.GetTrackLocal()
	slog.Debug("lobby.sessions: remove track", "trackId", track.ID(), "streamId", track.StreamID(), "sessionId", s.Id, "user", s.user)
	if s.sender != nil && s.sender.endpoint != nil {
		slog.Debug("lobby.sessions: removeTrack - from egress endpoint", "trackId", track.ID(), "streamId", track.StreamID(), "sessionId", s.Id, "user", s.user)
		s.sender.endpoint.RemoveTrack(track)
	}
}

type onInternallyQuit = func(ctx context.Context, user uuid.UUID) bool
