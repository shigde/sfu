package sessions

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/rtp"
	"github.com/shigde/sfu/internal/telemetry"
	"github.com/shigde/sfu/pkg/message"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/exp/slog"
)

const sessionTracer = telemetry.TracerName

type RtpEngine interface {
	EstablishEndpoint(ctx context.Context, sessionCtx context.Context, sessionId uuid.UUID, liveStream uuid.UUID, offer webrtc.SessionDescription, endpointType rtp.EndpointType, options ...rtp.EndpointOption) (*rtp.Endpoint, error)
	OfferEndpoint(ctx context.Context, sessionCtx context.Context, sessionId uuid.UUID, liveStream uuid.UUID, endpointType rtp.EndpointType, options ...rtp.EndpointOption) (*rtp.Endpoint, error)
}

var (
	ErrSessionAlreadyClosed         = errors.New("session already closed")
	ErrIngressAlreadyExists         = errors.New("ingress resource already exists in session")
	ErrEgressAlreadyExists          = errors.New("egress resource already exists in session")
	ErrNoSignalChannel              = errors.New("no signal channel connection exists in session")
	ErrSessionProcessWaitingTimeout = errors.New("session process waiting timeout")
	processWaitingTimeout           = 10 * time.Second // Ice gathering could take a long tine :-(

)

type Session struct {
	Id    uuid.UUID
	mutex sync.RWMutex
	ctx   context.Context
	user  uuid.UUID

	sessionType SessionType

	rtpEngine     RtpEngine
	hub           *Hub
	ingress       *rtp.Endpoint
	egress        *rtp.Endpoint
	signalChannel *rtp.Endpoint
	signal        *signal

	stop    context.CancelFunc
	garbage chan<- Item
}

func NewSession(ctx context.Context, user uuid.UUID, hub *Hub, engine RtpEngine, sType SessionType, garbage chan Item) *Session {
	sessionId := uuid.New()
	ctx = telemetry.ContextWithSessionValue(ctx, sessionId.String(), hub.LiveStreamId.String(), user.String())
	ctx, cancel := context.WithCancel(ctx)

	signal := newSignal(ctx, sessionId, user)

	session := &Session{
		Id:    sessionId,
		mutex: sync.RWMutex{},
		ctx:   ctx,
		user:  user,

		sessionType: sType,

		rtpEngine: engine,
		hub:       hub,
		signal:    signal,
		stop:      cancel,
		garbage:   garbage,
	}

	signal.onMuteCbk = session.onMuteTrack

	return session
}

// CreateIngressEndpoint
// Receives an Offer and generates an Answer to establish an Ingress WebRTC connection.
func (s *Session) CreateIngressEndpoint(ctx context.Context, offer *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	ctx, span := s.trace(ctx, "create_ingress_endpoint")
	defer span.End()
	if s.isDone() {
		return nil, telemetry.RecordError(span, ErrSessionAlreadyClosed)
	}

	if s.ingress != nil {
		return nil, telemetry.RecordError(span, ErrIngressAlreadyExists)
	}

	//signalSetting.OnDataC.

	option := make([]rtp.EndpointOption, 0)
	// option = append(option, rtp.EndpointWithDataChannel(setting.GetDataChannelCbk()))
	option = append(option, rtp.EndpointWithDataChannel(s.signal.OnIngressChannel))
	option = append(option, rtp.EndpointWithLostConnectionListener(s.onLostConnection))
	option = append(option, rtp.EndpointWithTrackDispatcher(s.hub))

	endpoint, err := s.rtpEngine.EstablishEndpoint(ctx, s.ctx, s.Id, s.hub.LiveStreamId, *offer, rtp.IngressEndpoint, option...)
	if err != nil {
		return nil, telemetry.RecordErrorf(span, "create rtp endpoint", err)
	}
	s.ingress = endpoint

	span.AddEvent("Wait for Local Description.")
	ctxTimeout, cancel := context.WithTimeout(ctx, processWaitingTimeout)
	defer cancel()
	answer, err := s.ingress.GetLocalDescription(ctxTimeout)
	if err != nil {
		return nil, telemetry.RecordErrorf(span, "create ingress answer resource", err)
	}

	span.AddEvent("Receive Local Description.")
	return answer, nil
}

// OfferIngressEndpoint
// Creates an offer and prepares an Ingress WebRTC connection. The connection then waits for the response from the
// other peer.
func (s *Session) OfferIngressEndpoint(ctx context.Context) (*webrtc.SessionDescription, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	ctx, span := s.trace(ctx, "create_ingress_endpoint")
	defer span.End()
	if s.isDone() {
		return nil, telemetry.RecordError(span, ErrSessionAlreadyClosed)
	}

	if s.ingress != nil {
		return nil, telemetry.RecordError(span, ErrIngressAlreadyExists)
	}

	//signalSetting.OnDataC.

	option := make([]rtp.EndpointOption, 0)
	// option = append(option, rtp.EndpointWithDataChannel(setting.GetDataChannelCbk()))
	option = append(option, rtp.EndpointWithDataChannel(s.signal.OnIngressChannel))
	option = append(option, rtp.EndpointWithLostConnectionListener(s.onLostConnection))
	option = append(option, rtp.EndpointWithTrackDispatcher(s.hub))

	endpoint, err := s.rtpEngine.OfferEndpoint(ctx, s.ctx, s.Id, s.hub.LiveStreamId, rtp.IngressEndpoint, option...)
	if err != nil {
		return nil, telemetry.RecordErrorf(span, "create rtp endpoint", err)
	}
	s.ingress = endpoint

	span.AddEvent("Wait for Local Description.")
	ctxTimeout, cancel := context.WithTimeout(ctx, processWaitingTimeout)
	defer cancel()
	answer, err := s.ingress.GetLocalDescription(ctxTimeout)
	if err != nil {
		return nil, telemetry.RecordErrorf(span, "create ingress answer resource", err)
	}

	span.AddEvent("Receive Local Description.")
	return answer, nil
}

// CreateEgressEndpoint
// Receives an Offer and generates an Answer to establish an Egress WebRTC connection.
func (s *Session) CreateEgressEndpoint(ctx context.Context, offer *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	ctx, span := s.trace(ctx, "create_egress_endpoint")
	defer span.End()
	if s.isDone() {
		return nil, telemetry.RecordError(span, ErrSessionAlreadyClosed)
	}

	if s.egress != nil {
		return nil, telemetry.RecordError(span, ErrEgressAlreadyExists)
	}

	// For an egress endpoint we need a data channel. The data channel is used for media update signaling.
	// That's why we're waiting until it's built
	// ----> data channel setup
	// signalOption.setUpSignalChannel(s.signal)

	if s.ingress == nil {
		return nil, telemetry.RecordError(span, ErrNoSignalChannel)
	}

	select {
	case err := <-s.signal.waitForMessengerSetupFinished():
		if err != nil {
			return nil, telemetry.RecordErrorf(span, "waiting for messenger", err)
		}
	case <-s.ctx.Done():
		return nil, telemetry.RecordError(span, ErrSessionAlreadyClosed)
	case <-time.After(processWaitingTimeout):
		return nil, telemetry.RecordError(span, ErrSessionProcessWaitingTimeout)
	}
	// <-- end date channel setup

	hub := s.hub
	withTrackCbk := rtp.EndpointWithGetCurrentTrackCbk(func(ctx context.Context, sessionId uuid.UUID) ([]*rtp.TrackInfo, error) {
		return hub.getTrackList(ctx, sessionId, filterForSession(sessionId))
	})

	option := make([]rtp.EndpointOption, 0)
	option = append(option, withTrackCbk)
	option = append(option, rtp.EndpointWithDataChannel(s.signal.OnEmptyChannel))
	option = append(option, rtp.EndpointWithNegotiationNeededListener(s.signal.OnNegotiationNeeded))
	option = append(option, rtp.EndpointWithLostConnectionListener(s.onLostConnection))

	endpoint, err := s.rtpEngine.EstablishEndpoint(ctx, s.ctx, s.Id, s.hub.LiveStreamId, *offer, rtp.EgressEndpoint, option...)
	if err != nil {
		return nil, telemetry.RecordErrorf(span, "create rtp endpoint", err)
	}
	s.egress = endpoint
	// @TODO: The signaling should be independent of ingress and egress
	s.signal.addEgressEndpoint(s.egress)

	span.AddEvent("Wait for Local Description.")
	ctxTimeout, cancel := context.WithTimeout(ctx, processWaitingTimeout)
	defer cancel()
	answer, err := s.egress.GetLocalDescription(ctxTimeout)
	if err != nil {
		return nil, telemetry.RecordErrorf(span, "create egress answer resource", err)
	}
	span.AddEvent("Receive Local Description.")
	return answer, nil
}

// OfferEgressEndpoint
// Creates an offer and prepares an Egress WebRTC connection. The connection then waits for the response from the
// other peer.
func (s *Session) OfferEgressEndpoint(ctx context.Context) (*webrtc.SessionDescription, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	ctx, span := s.trace(ctx, "offer_egress_endpoint")
	defer span.End()
	if s.isDone() {
		return nil, telemetry.RecordError(span, ErrSessionAlreadyClosed)
	}

	if s.egress != nil {
		return nil, telemetry.RecordError(span, ErrEgressAlreadyExists)
	}

	// For an egress endpoint we need a data channel. The data channel is used for media update signaling.
	// That's why we're waiting until it's built
	// ----> data channel setup
	// signalOption.setUpSignalChannel(s.signal)

	if s.ingress == nil {
		return nil, telemetry.RecordError(span, ErrNoSignalChannel)
	}

	select {
	case err := <-s.signal.waitForMessengerSetupFinished():
		if err != nil {
			return nil, telemetry.RecordErrorf(span, "waiting for messenger", err)
		}
	case <-s.ctx.Done():
		return nil, telemetry.RecordError(span, ErrSessionAlreadyClosed)
	case <-time.After(processWaitingTimeout):
		return nil, telemetry.RecordError(span, ErrSessionProcessWaitingTimeout)
	}
	// <-- end date channel setup

	hub := s.hub
	withTrackCbk := rtp.EndpointWithGetCurrentTrackCbk(func(ctx context.Context, sessionId uuid.UUID) ([]*rtp.TrackInfo, error) {
		return hub.getTrackList(ctx, sessionId, filterForSession(sessionId))
	})

	option := make([]rtp.EndpointOption, 0)
	option = append(option, withTrackCbk)
	option = append(option, rtp.EndpointWithDataChannel(s.signal.OnEmptyChannel))
	option = append(option, rtp.EndpointWithNegotiationNeededListener(s.signal.OnNegotiationNeeded))
	option = append(option, rtp.EndpointWithLostConnectionListener(s.onLostConnection))

	endpoint, err := s.rtpEngine.OfferEndpoint(ctx, s.ctx, s.Id, s.hub.LiveStreamId, rtp.EgressEndpoint, option...)
	if err != nil {
		return nil, telemetry.RecordErrorf(span, "create rtp endpoint", err)
	}
	s.egress = endpoint
	// @TODO: The signaling should be independent of ingress and egress
	s.signal.addEgressEndpoint(s.egress)

	span.AddEvent("Wait for Local Description.")
	ctxTimeout, cancel := context.WithTimeout(ctx, processWaitingTimeout)
	defer cancel()
	answer, err := s.egress.GetLocalDescription(ctxTimeout)
	if err != nil {
		return nil, telemetry.RecordErrorf(span, "create egress answer resource", err)
	}
	span.AddEvent("Receive Local Description.")
	return answer, nil
}

func (s *Session) onLostConnection() {
	slog.Warn("session: connect lost connection", "sessionId", s.Id, "userId", s.user)
	go func(session *Session) {
		select {
		case <-session.ctx.Done():
			slog.Debug("sessions: internally quit interrupted because session already closed", "session id", session.Id, "user", session.user)
		default:
			item := NewItem(s.user)
			select {
			case session.garbage <- item:
				session.stop()
				<-item.Done
				slog.Debug("sessions: internally stop of session", "session id", session.Id, "user", session.user)
			case <-session.ctx.Done():
				slog.Debug("sessions: internally stop interrupted because session already closed", "session id", session.Id, "user", session.user)
			}
		}
	}(s)
}

func (s *Session) addTrack(ctx context.Context, trackInfo *rtp.TrackInfo) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	track := trackInfo.GetTrackLocal()
	slog.Debug("sessions: add track", "trackId", track.ID(), "streamId", track.StreamID(), "sessionId", s.Id, "user", s.user)
	if s.egress != nil {
		slog.Debug("sessions: add track - to egress egress", "trackId", track.ID(), "streamId", track.StreamID(), "sessionId", s.Id, "user", s.user)
		ctx, span := s.trace(ctx, "egress_add_track")
		s.egress.AddTrack(ctx, trackInfo)
		span.End()
	}
}

func (s *Session) removeTrack(ctx context.Context, trackInfo *rtp.TrackInfo) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	track := trackInfo.GetTrackLocal()
	slog.Debug("bug-1: remove", "session", s.Id, "trackId", track.ID(), "kind", track.Kind())
	slog.Debug("sessions: remove track", "trackId", track.ID(), "streamId", track.StreamID(), "sessionId", s.Id, "user", s.user)
	if s.egress != nil {
		slog.Debug("sessions: removeTrack - from egress egress", "trackId", track.ID(), "streamId", track.StreamID(), "sessionId", s.Id, "user", s.user)
		ctx, span := s.trace(ctx, "egress_remove_track")
		s.egress.RemoveTrack(ctx, trackInfo)
		span.End()
	}
}

func (s *Session) muteTrack(ctx context.Context, trackInfo *rtp.TrackInfo) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// track telemetry data
	ctx, span := s.trace(ctx, "egress_mute_track_event")
	defer span.End()
	// telemetry
	span.SetAttributes(
		attribute.String("ingress_mid", trackInfo.IngressMid),
		attribute.String("ingress_mute", strconv.FormatBool(trackInfo.GetMute())),
	)

	if s.egress == nil {
		return
	}

	if egressInfo, ok := s.egress.SetEgressMute(trackInfo.GetId(), trackInfo.GetMute()); ok {
		// send event to client
		_ = s.signal.messenger.SendMute(&message.Mute{
			Mid:  egressInfo.GetEgressMid(),
			Mute: egressInfo.GetMute(),
		})
		// telemetry
		span.SetAttributes(
			attribute.String("egress_mid", egressInfo.GetEgressMid()),
			attribute.String("egress_mute", strconv.FormatBool(egressInfo.GetMute())),
		)
		// telemetry event
		span.AddEvent("Send Egress Mute to Client")
	}
}

func (s *Session) onMuteTrack(mute *message.Mute) {
	// track telemetry data
	ctx, span := s.trace(context.Background(), "ingress_mute_track_event")
	defer span.End()
	span.SetAttributes(
		attribute.String("mid", mute.Mid),
		attribute.String("mid", strconv.FormatBool(mute.Mute)),
	)
	if trackInfo, ok := s.ingress.SetIngressMute(mute.Mid, mute.Mute); ok {
		go s.hub.DispatchMuteTrack(ctx, trackInfo)
		// telemetry event
		span.AddEvent("Dispatch Mute to Sessions")
	}
}

func (s *Session) isDone() bool {
	select {
	case <-s.ctx.Done():
		return true
	default:
		return false
	}
}

func (s *Session) getType() SessionType {
	return s.sessionType
}

func (s *Session) trace(ctx context.Context, spanName string) (context.Context, trace.Span) {
	return telemetry.NewTraceSpan(ctx, s.ctx, "session: "+spanName)
}

func (s *Session) initComplete() bool {
	if looked := s.mutex.TryRLock(); !looked {
		return false
	}
	defer s.mutex.RUnlock()

	if s.egress != nil {
		return s.egress.IsInitComplete()
	}
	return false
}
