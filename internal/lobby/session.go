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
	Id        uuid.UUID
	ctx       context.Context
	user      uuid.UUID
	rtpEngine rtpEngine
	hub       *hub
	ingress   *rtp.Endpoint
	egress    *rtp.Endpoint
	signal    *signal
	reqChan   chan *sessionRequest
	doStop    chan<- uuid.UUID
	stop      context.CancelFunc
}

func newSession(user uuid.UUID, hub *hub, engine rtpEngine, doStop chan<- uuid.UUID) *session {
	ctx, cancel := context.WithCancel(context.Background())
	requests := make(chan *sessionRequest)
	sessionId := uuid.New()
	signal := newSignal(ctx, sessionId, user)

	session := &session{
		Id:        uuid.New(),
		ctx:       ctx,
		user:      user,
		rtpEngine: engine,
		hub:       hub,
		signal:    signal,
		reqChan:   requests,
		doStop:    doStop,
		stop:      cancel,
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
		case <-s.ctx.Done():
			// @TODO Take care that's every stream is closed!
			slog.Info("lobby.sessions: stop running", "session id", s.Id, "user", s.user)
			return
		}
	}
}

func (s *session) runRequest(req *sessionRequest) {
	slog.Debug("lobby.sessions: runRequest", "id", s.Id, "user", s.user)
	select {
	case s.reqChan <- req:
		slog.Debug("lobby.sessions: runRequest - return response", "id", s.Id, "user", s.user)
	case <-s.ctx.Done():
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

	if s.ingress != nil {
		return nil, errReceiverInSessionAlreadyExists
	}

	option := make([]rtp.EndpointOption, 0)
	option = append(option, rtp.EndpointWithDataChannel(s.signal.OnIngressChannel))
	option = append(option, rtp.EndpointWithLostConnectionListener(s.onLostConnectionListener))
	option = append(option, rtp.EndpointWithTrackDispatcher(s.hub))

	endpoint, err := s.rtpEngine.EstablishIngressEndpoint(s.ctx, s.Id, s.hub.LiveStreamId, *req.reqSDP, option...)
	if err != nil {
		return nil, fmt.Errorf("create rtp connection: %w", err)
	}
	s.ingress = endpoint
	ctxTimeout, cancel := context.WithTimeout(ctx, iceGatheringTimeout)
	defer cancel()
	answer, err := s.ingress.GetLocalDescription(ctxTimeout)
	if err != nil {
		return nil, fmt.Errorf("create rtp answerEgressReq: %w", err)
	}
	return answer, nil
}

func (s *session) handleAnswerEgressReq(req *sessionRequest) (*webrtc.SessionDescription, error) {
	_, span := otel.Tracer(tracerName).Start(req.ctx, "session:handleAnswerEgressReq")
	defer span.End()

	if s.egress == nil || s.signal.egressEndpoint == nil {
		return nil, errNoSenderInSession
	}

	s.signal.onAnswer(req.reqSDP, 0)
	return nil, nil
}

func (s *session) handleInitEgressReq(req *sessionRequest) (*webrtc.SessionDescription, error) {
	ctx, span := otel.Tracer(tracerName).Start(req.ctx, "session:handleInitEgressReq")
	defer span.End()

	if s.egress != nil {
		return nil, errSenderInSessionAlreadyExists
	}

	if s.ingress == nil {
		return nil, errNoReceiverInSession
	}

	select {
	case err := <-s.signal.waitForIngressDataChannel():
		if err != nil {
			return nil, fmt.Errorf("waiting for messenger: %w", err)
		}
	case <-s.ctx.Done():
		return nil, errSessionAlreadyClosed
	case <-time.After(sessionReqTimeout):
		return nil, errSessionRequestTimeout
	}

	hub := s.hub
	withTrackCbk := rtp.EndpointWithGetCurrentTrackCbk(func(sessionId uuid.UUID) ([]*rtp.TrackInfo, error) {
		return hub.getTrackList(sessionId, filterForSession(sessionId))
	})

	option := make([]rtp.EndpointOption, 0)
	option = append(option, withTrackCbk)
	option = append(option, rtp.EndpointWithDataChannel(s.signal.OnEgressChannel))
	option = append(option, rtp.EndpointWithNegotiationNeededListener(s.signal.OnNegotiationNeeded))
	option = append(option, rtp.EndpointWithLostConnectionListener(s.onLostConnectionListener))

	endpoint, err := s.rtpEngine.EstablishEgressEndpoint(s.ctx, s.Id, s.hub.LiveStreamId, option...)
	if err != nil {
		return nil, fmt.Errorf("create rtp connection: %w", err)
	}
	s.egress = endpoint
	s.signal.addEgressEndpoint(s.egress)

	ctxTimeout, cancel := context.WithTimeout(ctx, iceGatheringTimeout)
	defer cancel()
	offer, err := s.egress.GetLocalDescription(ctxTimeout)
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

	hub := s.hub
	withTrackCbk := rtp.EndpointWithGetCurrentTrackCbk(func(sessionId uuid.UUID) ([]*rtp.TrackInfo, error) {
		return hub.getTrackList(sessionId, filterForSession(sessionId))
	})

	option := make([]rtp.EndpointOption, 0)
	option = append(option, withTrackCbk)
	endpoint, err := s.rtpEngine.EstablishStaticEgressEndpoint(ctx, s.Id, s.hub.LiveStreamId, *req.reqSDP, option...)

	if err != nil {
		return nil, fmt.Errorf("create rtp connection: %w", err)
	}
	s.egress = endpoint
	ctxTimeout, cancel := context.WithTimeout(ctx, iceGatheringTimeout)
	defer cancel()
	answer, err := s.egress.GetLocalDescription(ctxTimeout)
	if err != nil {
		return nil, fmt.Errorf("create rtp answerEgressReq: %w", err)
	}
	return answer, nil
}

func (s *session) addTrack(trackInfo *rtp.TrackInfo) {
	track := trackInfo.GetTrackLocal()
	slog.Debug("lobby.sessions: add track", "trackId", track.ID(), "streamId", track.StreamID(), "sessionId", s.Id, "user", s.user)
	if s.egress != nil {
		slog.Debug("lobby.sessions: add track - to egress endpoint", "trackId", track.ID(), "streamId", track.StreamID(), "sessionId", s.Id, "user", s.user)
		s.egress.AddTrack(trackInfo)
	}
}

func (s *session) removeTrack(trackInfo *rtp.TrackInfo) {
	track := trackInfo.GetTrackLocal()
	slog.Debug("lobby.sessions: remove track", "trackId", track.ID(), "streamId", track.StreamID(), "sessionId", s.Id, "user", s.user)
	if s.egress != nil {
		slog.Debug("lobby.sessions: removeTrack - from egress endpoint", "trackId", track.ID(), "streamId", track.StreamID(), "sessionId", s.Id, "user", s.user)
		s.egress.RemoveTrack(trackInfo)
	}
}
func (s *session) onLostConnectionListener() {
	slog.Warn("lobby.session: endpoint lost connection", "sessionId", s.Id, "user", s.user)
	go func() {
		select {
		case <-s.ctx.Done():
			slog.Debug("lobby.sessions: internally quit interrupted because session already closed", "session id", s.Id, "user", s.user)
		default:
			select {
			case s.doStop <- s.user:
				slog.Debug("lobby.sessions: internally stop of session", "session id", s.Id, "user", s.user)
			case <-s.ctx.Done():
				slog.Debug("lobby.sessions: internally stop interrupted because session already closed", "session id", s.Id, "user", s.user)
			}
		}
	}()

}
