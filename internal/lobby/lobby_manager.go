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
	lobbies      *lobbyRepository
	lobbyGarbage chan<- lobbyItem
}

func NewLobbyManager(storage storage.Storage, e sessions.RtpEngine, homeUrl *url.URL, registerToken string) *LobbyManager {
	lobbyRep := newLobbyRepository(storage, e, homeUrl, registerToken)
	lobbyGarbage := make(chan lobbyItem)

	go func() {
		// if a lobby has no sessions anymore, the lobby triggers a delete process.
		for item := range lobbyGarbage {
			ok := lobbyRep.delete(context.Background(), item.LobbyId)
			if !ok {
				slog.Warn("lobby could not delete", "lobby", item.LobbyId)
			}
			item.Done <- ok
		}
	}()
	return &LobbyManager{lobbyRep, lobbyGarbage}
}

func (m *LobbyManager) NewIngressResource(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription, option ...resources.Option) (*resources.WebRTC, error) {
	lobbyObj, err := m.lobbies.getOrCreateLobby(ctx, lobbyId, m.lobbyGarbage)
	if err != nil {
		return nil, fmt.Errorf("getting or creating lobby: %w", err)
	}
	if ok := lobbyObj.newSession(user, sessions.UserSession); !ok {
		return nil, fmt.Errorf("creating new session failes")
	}

	cmd := commands.NewAnswerUserIngress(ctx, user, offer)
	lobbyObj.runCommand(cmd)

	select {
	case <-cmd.Done():
		return cmd.Response, cmd.Err
	case <-ctx.Done():
		return nil, fmt.Errorf("time out")
	}
}

func (m *LobbyManager) NewEgressResource(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription, option ...resources.Option) (*resources.WebRTC, error) {
	lobbyObj, err := m.lobbies.getOrCreateLobby(ctx, lobbyId, m.lobbyGarbage)
	if err != nil {
		return nil, fmt.Errorf("getting or creating lobby: %w", err)
	}

	cmd := commands.NewAnswerUserEgress(ctx, user, offer)
	lobbyObj.runCommand(cmd)

	select {
	case <-cmd.Done():
		return cmd.Response, cmd.Err
	case <-ctx.Done():
		return nil, fmt.Errorf("time out")
	}
}

// Old API -----------------------------------

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
