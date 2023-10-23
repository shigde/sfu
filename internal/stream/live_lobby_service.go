package stream

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/lobby"
)

type LiveLobbyService struct {
	// @TODO Space should not have this properties. Redactor this. This make a lot of trouble
	lobbyManager *lobby.LobbyManager
	store        storage
}

func NewLiveLobbyService(store storage, lobbyManager *lobby.LobbyManager) *LiveLobbyService {
	return &LiveLobbyService{
		store:        store,
		lobbyManager: lobbyManager,
	}
}

func (s *LiveLobbyService) EnterLobby(ctx context.Context, sdp *webrtc.SessionDescription, stream *LiveStream, userId uuid.UUID) (*webrtc.SessionDescription, string, error) {
	var resource string
	resourceData, err := s.lobbyManager.AccessLobby(ctx, stream.Lobby.UUID, userId, sdp)
	if err != nil {
		return nil, resource, fmt.Errorf("accessing lobby: %w", err)
	}
	resource = resourceData.Resource.String()
	return resourceData.Answer, resource, nil
}

func (s *LiveLobbyService) StartListenLobby(ctx context.Context, stream *LiveStream, userId uuid.UUID) (*webrtc.SessionDescription, error) {
	resourceData, err := s.lobbyManager.StartListenLobby(ctx, stream.Lobby.UUID, userId)
	if err != nil {
		return nil, fmt.Errorf("start listening lobby: %w", err)
	}
	if !resourceData.Active {
		return nil, ErrLobbyNotActive
	}
	return resourceData.Offer, nil
}

func (s *LiveLobbyService) ListenLobby(ctx context.Context, offer *webrtc.SessionDescription, stream *LiveStream, userId uuid.UUID) (bool, error) {
	resourceData, err := s.lobbyManager.ListenLobby(ctx, stream.Lobby.UUID, userId, offer)
	if err != nil {
		return false, fmt.Errorf("listening lobby: %w", err)
	}
	if !resourceData.Active {
		return false, ErrLobbyNotActive
	}
	return resourceData.Active, nil
}

func (s *LiveLobbyService) LeaveLobby(ctx context.Context, stream *LiveStream, userId uuid.UUID) (bool, error) {
	left, err := s.lobbyManager.LeaveLobby(ctx, stream.Lobby.UUID, userId)
	if err != nil {
		return false, fmt.Errorf("leave lobby: %w", err)
	}
	return left, nil
}

func (s *LiveLobbyService) StartLiveStream(ctx context.Context, stream *LiveStream, streamInfo *LiveStreamInfo, userId uuid.UUID) error {
	if err := s.lobbyManager.StartLiveStream(ctx, stream.Lobby.UUID, streamInfo.StreamKey, streamInfo.Rtmp, userId); err != nil {
		return fmt.Errorf("start live stream: %w", err)
	}
	return nil
}
