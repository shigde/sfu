package stream

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

type LiveLobbyService struct {
	lobbyManager liveLobbyManager
	store        storage
}

func NewLiveLobbyService(store storage, lobbyManager liveLobbyManager) *LiveLobbyService {
	return &LiveLobbyService{
		store:        store,
		lobbyManager: lobbyManager,
	}
}

func (s *LiveLobbyService) CreateLobbyIngressEndpoint(ctx context.Context, sdp *webrtc.SessionDescription, stream *LiveStream, userId uuid.UUID) (*webrtc.SessionDescription, string, error) {
	resource, err := s.lobbyManager.NewIngressResource(ctx, stream.Lobby.UUID, userId, sdp)
	if err != nil {
		return nil, "---", fmt.Errorf("accessing lobby: %w", err)
	}
	return resource.SDP, "---", nil
}

func (s *LiveLobbyService) InitLobbyEgressEndpoint(ctx context.Context, stream *LiveStream, userId uuid.UUID) (*webrtc.SessionDescription, error) {
	resourceData, err := s.lobbyManager.InitLobbyEgressEndpoint(ctx, stream.Lobby.UUID, userId)
	if err != nil {
		return nil, fmt.Errorf("start listening lobby: %w", err)
	}
	if !resourceData.Active {
		return nil, ErrLobbyNotActive
	}
	return resourceData.Offer, nil
}

func (s *LiveLobbyService) FinalCreateLobbyEgressEndpoint(ctx context.Context, offer *webrtc.SessionDescription, stream *LiveStream, userId uuid.UUID) (bool, error) {
	resourceData, err := s.lobbyManager.FinalCreateLobbyEgressEndpoint(ctx, stream.Lobby.UUID, userId, offer)
	if err != nil {
		return false, fmt.Errorf("listening lobby: %w", err)
	}
	if !resourceData.Active {
		return false, ErrLobbyNotActive
	}
	return resourceData.Active, nil
}

func (s *LiveLobbyService) CreateMainStreamLobbyEgressEndpoint(ctx context.Context, offer *webrtc.SessionDescription, stream *LiveStream, userId uuid.UUID) (*webrtc.SessionDescription, error) {
	resourceData, err := s.lobbyManager.CreateMainStreamLobbyEgressEndpoint(ctx, stream.Lobby.UUID, userId, offer)
	if err != nil {
		return nil, fmt.Errorf("accessing lobby: %w", err)
	}
	return resourceData.Answer, nil
}

func (s *LiveLobbyService) LeaveLobby(ctx context.Context, stream *LiveStream, userId uuid.UUID) (bool, error) {
	left, err := s.lobbyManager.LeaveLobby(ctx, stream.Lobby.UUID, userId)
	if err != nil {
		return false, fmt.Errorf("leave lobby: %w", err)
	}
	return left, nil
}

func (s *LiveLobbyService) StartLiveStream(ctx context.Context, stream *LiveStream, streamInfo *LiveStreamInfo, userId uuid.UUID) error {
	if err := s.lobbyManager.StartLiveStream(ctx, stream.Lobby.UUID, streamInfo.StreamKey, streamInfo.RtmpUrl, userId); err != nil {
		return fmt.Errorf("start live stream: %w", err)
	}
	return nil
}

func (s *LiveLobbyService) StopLiveStream(ctx context.Context, stream *LiveStream, userId uuid.UUID) error {
	if err := s.lobbyManager.StopLiveStream(ctx, stream.Lobby.UUID, userId); err != nil {
		return fmt.Errorf("stop live stream: %w", err)
	}
	return nil
}

func (s *LiveLobbyService) CreateLobbyHostPipeConnection(ctx context.Context, offer *webrtc.SessionDescription, stream *LiveStream, instanceId uuid.UUID) (*webrtc.SessionDescription, string, error) {
	resourceData, err := s.lobbyManager.CreateLobbyHostPipe(ctx, stream.Lobby.UUID, offer, instanceId)
	if err != nil {
		return nil, "", fmt.Errorf("creating lobby host pipe connection: %w", err)
	}
	return resourceData.Answer, "", nil
}

func (s *LiveLobbyService) CreateLobbyHostIngressConnection(ctx context.Context, offer *webrtc.SessionDescription, stream *LiveStream, instanceId uuid.UUID) (*webrtc.SessionDescription, string, error) {
	resourceData, err := s.lobbyManager.CreateLobbyHostIngress(ctx, stream.Lobby.UUID, offer, instanceId)
	if err != nil {
		return nil, "", fmt.Errorf("creating lobby host ingress connection: %w", err)
	}
	return resourceData.Answer, "", nil
}

func (s *LiveLobbyService) CloseLobbyHostConnection(ctx context.Context, stream *LiveStream, instanceId uuid.UUID) (bool, error) {
	left, err := s.lobbyManager.CloseLobbyHostPipe(ctx, stream.Lobby.UUID, instanceId)
	if err != nil {
		return false, fmt.Errorf("closing host pipe: %w", err)
	}
	return left, nil
}
