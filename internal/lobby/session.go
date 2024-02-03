package lobby

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/rtp"
	"github.com/shigde/sfu/pkg/message"
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
	sessionReqTimeout = 5 * time.Second

	iceGatheringTimeout = 5 * time.Second
)

type session struct {
	Id        uuid.UUID
	isRemote  bool
	ctx       context.Context
	user      uuid.UUID
	rtpEngine rtpEngine
	hub       *hub
	ingress   *rtp.Endpoint
	egress    *rtp.Endpoint
	channel   *rtp.Endpoint
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

	signal.onMuteCbk = session.onMuteTrack
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
		slog.Debug("lobby.sessions: runRequest - added to request queue", "id", s.Id, "user", s.user)
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
		// Data Channels for Server to server
	case offerHostRemotePipeReq:
		sdp, err = s.handleOfferHostRemotePipeReq(req)
	case offerHostPipeReq:
		sdp, err = s.handleOfferHostPipeReq(req)
	case answerHostPipeReq:
		sdp, err = s.handleAnswerHostPipeReq(req)
		// Media Server to Server
	case offerHostIngressReq:
		sdp, err = s.handleOfferHostIngressReq(req) // Remote Ingress egress Trigger
	case offerHostEgressReq:
		sdp, err = s.handleOfferHostEgressReq(req)
	case answerHostEgressReq:
		sdp, err = s.handleAnswerHostEgressReq(req)
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

	if s.egress == nil || s.signal.egress == nil {
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
	case err := <-s.signal.waitForMessengerSetupFinished():
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
	option = append(option, rtp.EndpointWithDataChannel(s.signal.OnEmptyChannel))
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
	// Without a message handler this sender is only a placeholder for the egress

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
		slog.Debug("lobby.sessions: add track - to egress egress", "trackId", track.ID(), "streamId", track.StreamID(), "sessionId", s.Id, "user", s.user)
		s.egress.AddTrack(trackInfo)
	}
}

func (s *session) removeTrack(trackInfo *rtp.TrackInfo) {
	track := trackInfo.GetTrackLocal()
	slog.Debug("lobby.sessions: remove track", "trackId", track.ID(), "streamId", track.StreamID(), "sessionId", s.Id, "user", s.user)
	if s.egress != nil {
		slog.Debug("lobby.sessions: removeTrack - from egress egress", "trackId", track.ID(), "streamId", track.StreamID(), "sessionId", s.Id, "user", s.user)
		s.egress.RemoveTrack(trackInfo)
	}
}
func (s *session) onLostConnectionListener() {
	slog.Warn("lobby.session: egress lost connection", "sessionId", s.Id, "user", s.user)
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

// handle mute
func (s *session) onMuteTrack(mute *message.Mute) {
	if trackInfo, ok := s.ingress.SetIngressMute(mute.Mid, mute.Mute); ok {
		go s.hub.DispatchMuteTrack(trackInfo)
	}
}

func (s *session) sendMuteTrack(info *rtp.TrackInfo) {
	if s.egress == nil {
		return
	}

	if egressInfo, ok := s.egress.SetEgressMute(info.GetId(), info.GetMute()); ok {
		_ = s.signal.messenger.sendMute(&message.Mute{
			Mid:  egressInfo.GetEgressMid(),
			Mute: egressInfo.GetMute(),
		})
	}
}

func (s *session) handleOfferHostIngressReq(req *sessionRequest) (*webrtc.SessionDescription, error) {
	ctx, span := otel.Tracer(tracerName).Start(req.ctx, "session:handleOfferHostIngressReq")
	defer span.End()
	slog.Debug("lobby.sessions: handleOfferHostIngressReq", "id", s.Id, "instanceId", s.user)

	if s.ingress != nil {
		return nil, errReceiverInSessionAlreadyExists
	}

	option := make([]rtp.EndpointOption, 0)
	option = append(option, rtp.EndpointWithDataChannel(s.signal.OnEmptyChannel))
	option = append(option, rtp.EndpointWithLostConnectionListener(s.onLostConnectionListener))
	option = append(option, rtp.EndpointWithTrackDispatcher(s.hub))

	endpoint, err := s.rtpEngine.EstablishIngressEndpoint(s.ctx, s.Id, s.hub.LiveStreamId, *req.reqSDP, option...)
	if err != nil {
		return nil, fmt.Errorf("create rtp connection: %w", err)
	}
	s.ingress = endpoint
	go func() {
		select {
		case <-s.signal.receivedMessenger:
			s.signal.addIngressEndpoint(endpoint)
		case <-s.ctx.Done():
			return
		}
	}()

	ctxTimeout, cancel := context.WithTimeout(ctx, iceGatheringTimeout)
	defer cancel()
	answer, err := s.ingress.GetLocalDescription(ctxTimeout)
	if err != nil {
		return nil, fmt.Errorf("create rtp handleOfferHostIngressReq: %w", err)
	}
	return answer, nil
}

func (s *session) handleOfferHostEgressReq(req *sessionRequest) (*webrtc.SessionDescription, error) {
	ctx, span := otel.Tracer(tracerName).Start(req.ctx, "session:handleInitEgressReq")
	defer span.End()
	slog.Debug("lobby.sessions: handleOfferHostEgressReq", "id", s.Id, "instanceId", s.user)

	if s.egress != nil {
		return nil, errSenderInSessionAlreadyExists
	}

	hub := s.hub
	withTrackCbk := rtp.EndpointWithGetCurrentTrackCbk(func(sessionId uuid.UUID) ([]*rtp.TrackInfo, error) {
		return hub.getTrackList(sessionId, filterForSession(sessionId))
	})

	option := make([]rtp.EndpointOption, 0)
	option = append(option, withTrackCbk)
	option = append(option, rtp.EndpointWithDataChannel(s.signal.OnEmptyChannel))
	option = append(option, rtp.EndpointWithNegotiationNeededListener(s.signal.OnNegotiationNeeded))
	option = append(option, rtp.EndpointWithLostConnectionListener(s.onLostConnectionListener))

	endpoint, err := s.rtpEngine.EstablishEgressEndpoint(s.ctx, s.Id, s.hub.LiveStreamId, option...)
	if err != nil {
		return nil, fmt.Errorf("create rtp connection: %w", err)
	}
	s.egress = endpoint
	go func() {
		select {
		case <-s.signal.receivedMessenger:
			s.signal.addEgressEndpoint(s.egress)
		case <-s.ctx.Done():
			return
		}
	}()

	ctxTimeout, cancel := context.WithTimeout(ctx, iceGatheringTimeout)
	defer cancel()
	offer, err := s.egress.GetLocalDescription(ctxTimeout)
	if err != nil {
		return nil, fmt.Errorf("create rtp answerEgressReq: %w", err)
	}

	return offer, nil
}

func (s *session) handleAnswerHostEgressReq(req *sessionRequest) (*webrtc.SessionDescription, error) {
	_, span := otel.Tracer(tracerName).Start(req.ctx, "session:handleAnswerHostEgressReq")
	defer span.End()
	slog.Debug("lobby.sessions: handleAnswerHostEgressReq", "id", s.Id, "instanceId", s.user)

	if s.egress == nil || s.signal.egress == nil {
		return nil, errNoSenderInSession
	}

	s.signal.onAnswer(req.reqSDP, 0)
	return nil, nil
}

// Data Channel Endpoints
func (s *session) handleOfferHostRemotePipeReq(req *sessionRequest) (*webrtc.SessionDescription, error) {
	ctx, span := otel.Tracer(tracerName).Start(req.ctx, "session:handleOfferHostRemotePipeReq")
	defer span.End()
	slog.Debug("lobby.sessions: handleOfferHostRemotePipeReq", "id", s.Id, "instanceId", s.user)

	if s.channel != nil {
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
	s.channel = endpoint
	ctxTimeout, cancel := context.WithTimeout(ctx, iceGatheringTimeout)
	defer cancel()
	answer, err := s.channel.GetLocalDescription(ctxTimeout)
	if err != nil {
		return nil, fmt.Errorf("create rtp handleOfferHostRemotePipeReq: %w", err)
	}
	return answer, nil
}

func (s *session) handleOfferHostPipeReq(req *sessionRequest) (*webrtc.SessionDescription, error) {
	ctx, span := otel.Tracer(tracerName).Start(req.ctx, "session:handleOfferHostPipeReq")
	defer span.End()
	slog.Debug("lobby.sessions: handleOfferHostPipeReq", "id", s.Id, "instanceId", s.user)

	if s.channel != nil {
		return nil, errSenderInSessionAlreadyExists
	}

	option := make([]rtp.EndpointOption, 0)
	option = append(option, rtp.EndpointWithDataChannel(s.signal.OnHostPipeChannel))
	option = append(option, rtp.EndpointWithLostConnectionListener(s.onLostConnectionListener))

	endpoint, err := s.rtpEngine.EstablishEgressEndpoint(s.ctx, s.Id, s.hub.LiveStreamId, option...)
	if err != nil {
		return nil, fmt.Errorf("create rtp connection: %w", err)
	}
	s.channel = endpoint

	ctxTimeout, cancel := context.WithTimeout(ctx, iceGatheringTimeout)
	defer cancel()
	offer, err := s.channel.GetLocalDescription(ctxTimeout)
	if err != nil {
		return nil, fmt.Errorf("create rtp answerPipeReq: %w", err)
	}

	return offer, nil
}

func (s *session) handleAnswerHostPipeReq(req *sessionRequest) (*webrtc.SessionDescription, error) {
	_, span := otel.Tracer(tracerName).Start(req.ctx, "session:handleAnswerHostPipeReq")
	defer span.End()
	slog.Debug("lobby.sessions: handleAnswerHostPipeReq", "id", s.Id, "instanceId", s.user)

	if s.channel == nil {
		return nil, errNoSenderInSession
	}
	_ = s.channel.SetAnswer(req.reqSDP)
	return nil, nil
}
