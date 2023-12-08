package stream

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/metric"
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
	var resource string
	ctx = metric.ContextWithStream(ctx, stream.UUID.String())
	resourceData, err := s.lobbyManager.CreateLobbyIngressEndpoint(ctx, stream.Lobby.UUID, userId, sdp)
	if err != nil {
		return nil, resource, fmt.Errorf("accessing lobby: %w", err)
	}
	resource = resourceData.Resource.String()
	return resourceData.Answer, resource, nil
}

func (s *LiveLobbyService) InitLobbyEgressEndpoint(ctx context.Context, stream *LiveStream, userId uuid.UUID) (*webrtc.SessionDescription, error) {
	ctx = metric.ContextWithStream(ctx, stream.UUID.String())
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
	ctx = metric.ContextWithStream(ctx, stream.UUID.String())
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
	ctx = metric.ContextWithStream(ctx, stream.UUID.String())
	resourceData, err := s.lobbyManager.CreateMainStreamLobbyEgressEndpoint(ctx, stream.Lobby.UUID, userId, offer)
	if err != nil {
		return nil, fmt.Errorf("accessing lobby: %w", err)
	}
	return resourceData.Answer, nil
}

func (s *LiveLobbyService) LeaveLobby(ctx context.Context, stream *LiveStream, userId uuid.UUID) (bool, error) {
	ctx = metric.ContextWithStream(ctx, stream.UUID.String())
	left, err := s.lobbyManager.LeaveLobby(ctx, stream.Lobby.UUID, userId)
	if err != nil {
		return false, fmt.Errorf("leave lobby: %w", err)
	}
	return left, nil
}

func (s *LiveLobbyService) StartLiveStream(ctx context.Context, stream *LiveStream, streamInfo *LiveStreamInfo, userId uuid.UUID) error {
	ctx = metric.ContextWithStream(ctx, stream.UUID.String())
	if err := s.lobbyManager.StartLiveStream(ctx, stream.Lobby.UUID, streamInfo.StreamKey, streamInfo.RtmpUrl, userId); err != nil {
		return fmt.Errorf("start live stream: %w", err)
	}
	return nil
}

func (s *LiveLobbyService) StopLiveStream(ctx context.Context, stream *LiveStream, userId uuid.UUID) error {
	ctx = metric.ContextWithStream(ctx, stream.UUID.String())
	if err := s.lobbyManager.StopLiveStream(ctx, stream.Lobby.UUID, userId); err != nil {
		return fmt.Errorf("stop live stream: %w", err)
	}
	return nil
}
