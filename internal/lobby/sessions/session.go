package sessions

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/rtp"
	"golang.org/x/exp/slog"
)

type RtpEngine interface {
	EstablishIngressEndpoint(ctx context.Context, sessionId uuid.UUID, liveStream uuid.UUID, offer webrtc.SessionDescription, options ...rtp.EndpointOption) (*rtp.Endpoint, error)
	EstablishEgressEndpoint(ctx context.Context, sessionId uuid.UUID, liveStream uuid.UUID, options ...rtp.EndpointOption) (*rtp.Endpoint, error)
	EstablishStaticEgressEndpoint(ctx context.Context, sessionId uuid.UUID, liveStream uuid.UUID, offer webrtc.SessionDescription, options ...rtp.EndpointOption) (*rtp.Endpoint, error)
}

type Session struct {
	*sync.RWMutex
	Id       uuid.UUID
	user     uuid.UUID
	isRemote bool
	hub      *Hub

	rtpEngine     RtpEngine
	ingress       *rtp.Endpoint
	egress        *rtp.Endpoint
	signalChannel *rtp.Endpoint
	signal        *signal

	stop context.CancelFunc
	Done <-chan struct{}
}

func NewSession(ctx context.Context, user uuid.UUID, hub *Hub, engine RtpEngine) *Session {
	ctx, cancel := context.WithCancel(ctx)
	sessionId := uuid.New()
	signal := newSignal(ctx, sessionId, user)

	session := &Session{
		Id: uuid.New(),

		user:      user,
		rtpEngine: engine,
		hub:       hub,
		signal:    signal,
		stop:      cancel,
		Done:      ctx.Done(),
	}

	// signal.onMuteCbk = session.onMuteTrack
	go session.run(ctx)

	return session
}

func (s *Session) run(ctx context.Context) {
	slog.Info("lobby.sessions: run", "id", s.Id, "user", s.user)
	for {
		select {
		case <-ctx.Done():
			// @TODO Take care that's every stream is closed!
			slog.Info("lobby.sessions: stop running", "session id", s.Id, "user", s.user)
			return
		}
	}
}
