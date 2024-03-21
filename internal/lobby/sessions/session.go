package sessions

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/rtp"
	"github.com/shigde/sfu/internal/telemetry"
	"github.com/shigde/sfu/pkg/message"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/exp/slog"
)

const sessionTracer = telemetry.TracerName

type RtpEngine interface {
	EstablishEndpoint(ctx context.Context, sessionId uuid.UUID, liveStream uuid.UUID, offer webrtc.SessionDescription, endpointType rtp.EndpointType, options ...rtp.EndpointOption) (*rtp.Endpoint, error)
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
	Id       uuid.UUID
	mutex    sync.RWMutex
	ctx      context.Context
	user     uuid.UUID
	isRemote bool
	hub      *Hub

	rtpEngine     RtpEngine
	ingress       *rtp.Endpoint
	egress        *rtp.Endpoint
	signalChannel *rtp.Endpoint
	signal        *signal

	stop    context.CancelFunc
	garbage chan<- Item
}

func NewSession(ctx context.Context, user uuid.UUID, hub *Hub, engine RtpEngine, garbage chan Item) *Session {
	ctx, cancel := context.WithCancel(ctx)
	sessionId := uuid.New()
	signal := newSignal(ctx, sessionId, user)

	session := &Session{
		Id:    uuid.New(),
		mutex: sync.RWMutex{},
		ctx:   ctx,
		user:  user,

		rtpEngine: engine,
		hub:       hub,
		signal:    signal,
		stop:      cancel,
		garbage:   garbage,
	}

	signal.onMuteCbk = session.onMuteTrack

	return session
}

func (s *Session) CreateIngressEndpoint(ctx context.Context, offer *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	ctx, span := s.trace(ctx, "create_ingress_endpoint")
	defer span.End()

	if s.ingress != nil {
		return nil, telemetry.RecordError(span, ErrIngressAlreadyExists)
	}

	option := make([]rtp.EndpointOption, 0)
	option = append(option, rtp.EndpointWithDataChannel(s.signal.OnIngressChannel))
	option = append(option, rtp.EndpointWithLostConnectionListener(s.onLostConnection))
	option = append(option, rtp.EndpointWithTrackDispatcher(s.hub))

	endpoint, err := s.rtpEngine.EstablishEndpoint(s.ctx, s.Id, s.hub.LiveStreamId, *offer, rtp.IngressEndpoint, option...)
	if err != nil {
		return nil, telemetry.RecordErrorf(span, "create rtp connection", err)
	}
	s.ingress = endpoint

	span.AddEvent("Wait for Local Description.")
	ctxTimeout, cancel := context.WithTimeout(ctx, processWaitingTimeout)
	defer cancel()
	answer, err := s.ingress.GetLocalDescription(ctxTimeout)
	if err != nil {
		return nil, telemetry.RecordErrorf(span, "create ingress resource endpoint", err)
	}

	span.AddEvent("Receive Local Description.")
	return answer, nil
}

func (s *Session) CreateEgressEndpoint(ctx context.Context, offer *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	ctx, span := s.trace(ctx, "create_egress_endpoint")
	defer span.End()

	if s.egress != nil {
		return nil, telemetry.RecordError(span, ErrEgressAlreadyExists)
	}

	// For an egress endpoint we need a data channel. The data channel is used for media update signaling.
	// That's why we're waiting until it's built
	// ----> data channel setup
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
	withTrackCbk := rtp.EndpointWithGetCurrentTrackCbk(func(sessionId uuid.UUID) ([]*rtp.TrackInfo, error) {
		return hub.getTrackList(sessionId, filterForSession(sessionId))
	})

	option := make([]rtp.EndpointOption, 0)
	option = append(option, withTrackCbk)
	option = append(option, rtp.EndpointWithDataChannel(s.signal.OnEmptyChannel))
	option = append(option, rtp.EndpointWithNegotiationNeededListener(s.signal.OnNegotiationNeeded))
	option = append(option, rtp.EndpointWithLostConnectionListener(s.onLostConnection))

	endpoint, err := s.rtpEngine.EstablishEndpoint(s.ctx, s.Id, s.hub.LiveStreamId, *offer, rtp.EgressEndpoint, option...)
	if err != nil {
		return nil, telemetry.RecordErrorf(span, "create rtp connection", err)
	}
	s.egress = endpoint
	s.signal.addEgressEndpoint(s.egress)

	span.AddEvent("Wait for Local Description.")
	ctxTimeout, cancel := context.WithTimeout(ctx, processWaitingTimeout)
	defer cancel()
	answer, err := s.egress.GetLocalDescription(ctxTimeout)
	if err != nil {
		return nil, telemetry.RecordErrorf(span, "create egress resource endpoint", err)
	}
	span.AddEvent("Receive Local Description.")
	return answer, nil
}

func (s *Session) onLostConnection() {
	slog.Warn("session: connect lost connection", "sessionId", s.Id, "userId", s.user)
	go func() {
		select {
		case <-s.ctx.Done():
			slog.Debug("sessions: internally quit interrupted because session already closed", "session id", s.Id, "user", s.user)
		default:
			item := NewItem(s.user)
			select {
			case s.garbage <- item:
				s.stop()
				<-item.Done
				slog.Debug("sessions: internally stop of session", "session id", s.Id, "user", s.user)
			case <-s.ctx.Done():
				slog.Debug("sessions: internally stop interrupted because session already closed", "session id", s.Id, "user", s.user)
			}
		}
	}()
}

func (s *Session) addTrack(trackInfo *rtp.TrackInfo) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	track := trackInfo.GetTrackLocal()
	slog.Debug("sessions: add track", "trackId", track.ID(), "streamId", track.StreamID(), "sessionId", s.Id, "user", s.user)
	if s.egress != nil {
		slog.Debug("sessions: add track - to egress egress", "trackId", track.ID(), "streamId", track.StreamID(), "sessionId", s.Id, "user", s.user)
		s.egress.AddTrack(trackInfo)
	}
}

func (s *Session) removeTrack(trackInfo *rtp.TrackInfo) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	track := trackInfo.GetTrackLocal()
	slog.Debug("sessions: remove track", "trackId", track.ID(), "streamId", track.StreamID(), "sessionId", s.Id, "user", s.user)
	if s.egress != nil {
		slog.Debug("sessions: removeTrack - from egress egress", "trackId", track.ID(), "streamId", track.StreamID(), "sessionId", s.Id, "user", s.user)
		s.egress.RemoveTrack(trackInfo)
	}
}

func (s *Session) muteTrack(trackInfo *rtp.TrackInfo) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.egress == nil {
		return
	}

	if egressInfo, ok := s.egress.SetEgressMute(trackInfo.GetId(), trackInfo.GetMute()); ok {
		_ = s.signal.messenger.SendMute(&message.Mute{
			Mid:  egressInfo.GetEgressMid(),
			Mute: egressInfo.GetMute(),
		})
	}
}

func (s *Session) onMuteTrack(mute *message.Mute) {
	if trackInfo, ok := s.ingress.SetIngressMute(mute.Mid, mute.Mute); ok {
		go s.hub.DispatchMuteTrack(trackInfo)
	}
}

func (s *Session) trace(ctx context.Context, spanName string) (context.Context, trace.Span) {
	ctx, span := otel.Tracer(sessionTracer).Start(ctx, "session: "+spanName, trace.WithAttributes(
		attribute.String("sessionId", s.Id.String()),
		attribute.String("userId", s.user.String()),
	))
	return ctx, span
}
