package lobby

import (
	"context"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/lobby/commands"
	"github.com/shigde/sfu/internal/lobby/resources"
	"github.com/shigde/sfu/internal/lobby/sessions"
	"github.com/shigde/sfu/internal/storage"
	"golang.org/x/exp/slog"
)

type LobbyManager struct {
	lobbies               *lobbyRepository
	lobbyGarbageCollector chan<- uuid.UUID
}

func NewLobbyManager(storage storage.Storage, e sessions.RtpEngine, homeUrl *url.URL, registerToken string) *LobbyManager {
	lobbyRep := newLobbyRepository(storage, e, homeUrl, registerToken)
	lobbyGarbageCollector := make(chan uuid.UUID)
	go func() {
		for id := range lobbyGarbageCollector {
			if ok := lobbyRep.delete(context.Background(), id); !ok {
				slog.Warn("lobby could not delete", "lobby", id)
			}
		}
	}()
	return &LobbyManager{lobbyRep, lobbyGarbageCollector}
}

func (m *LobbyManager) NewIngressResource(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription, option ...resources.Option) (*resources.WebRTC, error) {
	_, err := m.lobbies.getOrCreateLobby(ctx, lobbyId, m.lobbyGarbageCollector)
	if err != nil {
		return nil, fmt.Errorf("getting or creating lobby: %w", err)
	}

	_ = commands.CreateIngressResourceCommand{}

	return nil, nil
}

//	NewEgressResource(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID, option WebrtcResourceOption) (*WebrtcResource, error)
//	DeleteAllResources(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID) (bool, error)
//}

//
//	NewIngressEndpoint(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription) (struct {
//		Answer       *webrtc.SessionDescription
//		Resource     uuid.UUID
//		RtpSessionId uuid.UUID
//	}, error)
//
//	InitLobbyEgressEndpoint(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID) (struct {
//		Offer        *webrtc.SessionDescription
//		Active       bool
//		RtpSessionId uuid.UUID
//	}, error)
//
//	FinalCreateLobbyEgressEndpoint(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription) (struct {
//		Answer       *webrtc.SessionDescription
//		Active       bool
//		RtpSessionId uuid.UUID
//	}, error)
//
//	CreateMainStreamLobbyEgressEndpoint(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription) (struct {
//		Answer       *webrtc.SessionDescription
//		RtpSessionId uuid.UUID
//	}, error)
//
//	LeaveLobby(ctx context.Context, lobbyId uuid.UUID, userId uuid.UUID) (bool, error)
//
//	// Live Stream API
//	StartLiveStream(
//		ctx context.Context,
//		lobbyId uuid.UUID,
//		key string,
//		rtmpUrl string,
//		userId uuid.UUID,
//	) error
//
//	StopLiveStream(
//		ctx context.Context,
//		lobbyId uuid.UUID,
//		userId uuid.UUID,
//	) error
//
//	// Server to Server API
//	CreateLobbyHostPipe(ctx context.Context, u uuid.UUID, offer *webrtc.SessionDescription, instanceId uuid.UUID) (struct {
//		Answer       *webrtc.SessionDescription
//		Resource     uuid.UUID
//		RtpSessionId uuid.UUID
//	}, error)
//
//	CreateLobbyHostIngress(ctx context.Context, u uuid.UUID, offer *webrtc.SessionDescription, instanceId uuid.UUID) (struct {
//		Answer       *webrtc.SessionDescription
//		Resource     uuid.UUID
//		RtpSessionId uuid.UUID
//	}, error)
//
//	CloseLobbyHostPipe(ctx context.Context, u uuid.UUID, id uuid.UUID) (bool, error)
//}

func (m *LobbyManager) CreateLobbyIngressEndpoint(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription) (struct {
	Answer       *webrtc.SessionDescription
	Resource     uuid.UUID
	RtpSessionId uuid.UUID
}, error) {
	return struct {
		Answer       *webrtc.SessionDescription
		Resource     uuid.UUID
		RtpSessionId uuid.UUID
	}{Answer: nil, Resource: uuid.New(), RtpSessionId: uuid.New()}, nil
}

func (m *LobbyManager) InitLobbyEgressEndpoint(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID) (struct {
	Offer        *webrtc.SessionDescription
	Active       bool
	RtpSessionId uuid.UUID
}, error) {
	return struct {
		Offer        *webrtc.SessionDescription
		Active       bool
		RtpSessionId uuid.UUID
	}{Offer: nil, Active: false, RtpSessionId: uuid.New()}, nil
}

func (m *LobbyManager) FinalCreateLobbyEgressEndpoint(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription) (struct {
	Answer       *webrtc.SessionDescription
	Active       bool
	RtpSessionId uuid.UUID
}, error) {
	return struct {
		Answer       *webrtc.SessionDescription
		Active       bool
		RtpSessionId uuid.UUID
	}{Answer: nil, Active: false, RtpSessionId: uuid.New()}, nil
}

func (m *LobbyManager) CreateMainStreamLobbyEgressEndpoint(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription) (struct {
	Answer       *webrtc.SessionDescription
	RtpSessionId uuid.UUID
}, error) {
	return struct {
		Answer       *webrtc.SessionDescription
		RtpSessionId uuid.UUID
	}{Answer: nil, RtpSessionId: uuid.New()}, nil
}

func (m *LobbyManager) LeaveLobby(ctx context.Context, lobbyId uuid.UUID, userId uuid.UUID) (bool, error) {
	return false, nil
}

// Live Stream API
func (m *LobbyManager) StartLiveStream(
	ctx context.Context,
	lobbyId uuid.UUID,
	key string,
	rtmpUrl string,
	userId uuid.UUID,
) error {
	return nil
}

func (m *LobbyManager) StopLiveStream(
	ctx context.Context,
	lobbyId uuid.UUID,
	userId uuid.UUID,
) error {
	return nil
}

// Server to Server API
func (m *LobbyManager) CreateLobbyHostPipe(ctx context.Context, u uuid.UUID, offer *webrtc.SessionDescription, instanceId uuid.UUID) (struct {
	Answer       *webrtc.SessionDescription
	Resource     uuid.UUID
	RtpSessionId uuid.UUID
}, error) {
	return struct {
		Answer       *webrtc.SessionDescription
		Resource     uuid.UUID
		RtpSessionId uuid.UUID
	}{Answer: nil, Resource: uuid.New(), RtpSessionId: uuid.New()}, nil
}

func (m *LobbyManager) CreateLobbyHostIngress(ctx context.Context, u uuid.UUID, offer *webrtc.SessionDescription, instanceId uuid.UUID) (struct {
	Answer       *webrtc.SessionDescription
	Resource     uuid.UUID
	RtpSessionId uuid.UUID
}, error) {
	return struct {
		Answer       *webrtc.SessionDescription
		Resource     uuid.UUID
		RtpSessionId uuid.UUID
	}{Answer: nil, Resource: uuid.New(), RtpSessionId: uuid.New()}, nil
}

func (m *LobbyManager) CloseLobbyHostPipe(ctx context.Context, u uuid.UUID, id uuid.UUID) (bool, error) {
	return false, nil
}
