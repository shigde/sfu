package stream

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

type lobbyListenAccessor interface {
	AccessLobby(ctx context.Context, liveStreamId uuid.UUID, userId uuid.UUID, offer *webrtc.SessionDescription) (struct {
		Answer       *webrtc.SessionDescription
		Resource     uuid.UUID
		RtpSessionId uuid.UUID
	}, error)

	ListenLobby(ctx context.Context, liveStreamId uuid.UUID, userId uuid.UUID, offer *webrtc.SessionDescription) (struct {
		Answer       *webrtc.SessionDescription
		Resource     uuid.UUID
		RtpSessionId uuid.UUID
	}, error)
}

type Space struct {
	Id             string                `json:"Id" gorm:"primaryKey"`
	LiveStreamRepo *LiveStreamRepository `gorm:"-"`
	lobby          lobbyListenAccessor   `gorm:"-"`
	store          storage               `gorm:"-"`
	entity
}

func newSpace(id string, lobby lobbyListenAccessor, store storage) (*Space, error) {
	repo, err := NewLiveStreamRepository(store)
	if err != nil {
		return nil, fmt.Errorf("creating live stream repository")
	}
	return &Space{Id: id, LiveStreamRepo: repo, lobby: lobby, store: store}, nil
}

func (s *Space) EnterLobby(ctx context.Context, sdp *webrtc.SessionDescription, stream *LiveStream, userId uuid.UUID) (*webrtc.SessionDescription, string, error) {
	var resource string
	resourceData, err := s.lobby.AccessLobby(ctx, stream.UUID, userId, sdp)
	if err != nil {
		return nil, resource, fmt.Errorf("accessing lobby: %w", err)
	}
	resource = resourceData.Resource.String()
	return resourceData.Answer, resource, nil
}

func (s *Space) ListenLobby(ctx context.Context, offer *webrtc.SessionDescription, stream *LiveStream, id uuid.UUID) (*webrtc.SessionDescription, string, error) {
	var resource string
	resourceData, err := s.lobby.ListenLobby(ctx, stream.UUID, id, offer)
	if err != nil {
		return nil, resource, fmt.Errorf("accessing lobby: %w", err)
	}
	resource = resourceData.Resource.String()
	return resourceData.Answer, resource, nil
}
