package stream

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/pkg/lobby"
)

type lobbyAccessor interface {
	AccessLobby(id string) (*lobby.RtpStreamLobby, error)
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
	lobbySpace, err := s.lobby.AccessLobby(stream.Id)
	var resource string
	if err != nil {
		return nil, resource, fmt.Errorf("creating lobby: %w", err)
	}
	resourceData, err := lobbySpace.Join(userId, sdp)
	if err != nil {
		return nil, resource, fmt.Errorf("joining lobby: %w", err)
	}
	resource = resourceData.Resource.String()

	return resourceData.Answer, resource, nil
}
