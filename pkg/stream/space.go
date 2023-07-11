package stream

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

type lobbyAccessor interface {
	AccessLobby(liveStreamId uuid.UUID, userId uuid.UUID, offer *webrtc.SessionDescription) (struct {
		Answer       *webrtc.SessionDescription
		Resource     uuid.UUID
		RtpSessionId uuid.UUID
	}, error)
}

type Space struct {
	Id             string                `json:"Id" gorm:"primaryKey"`
	LiveStreamRepo *LiveStreamRepository `gorm:"-"`
	lobby          lobbyAccessor         `gorm:"-"`
	store          storage               `gorm:"-"`
	entity
}

func newSpace(id string, lobby lobbyAccessor, store storage) (*Space, error) {
	repo, err := NewLiveStreamRepository(store)
	if err != nil {
		return nil, fmt.Errorf("creating live stream repository")
	}
	return &Space{Id: id, LiveStreamRepo: repo, lobby: lobby, store: store}, nil
}

func (s *Space) EnterLobby(sdp *webrtc.SessionDescription, stream *LiveStream, userId uuid.UUID) (*webrtc.SessionDescription, string, error) {
	var resource string
	resourceData, err := s.lobby.AccessLobby(stream.UUID, userId, sdp)
	if err != nil {
		return nil, resource, fmt.Errorf("accessing lobby: %w", err)
	}
	resource = resourceData.Resource.String()
	return resourceData.Answer, resource, nil
}
