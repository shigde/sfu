package stream

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/auth"
)

var ErrLobbyNotActive = errors.New("lobby not active")

type lobbyListenAccessor interface {
	AccessLobby(ctx context.Context, liveStreamId uuid.UUID, userId uuid.UUID, offer *webrtc.SessionDescription) (struct {
		Answer       *webrtc.SessionDescription
		Resource     uuid.UUID
		RtpSessionId uuid.UUID
	}, error)

	StartListenLobby(ctx context.Context, liveStreamId uuid.UUID, userId uuid.UUID) (struct {
		Offer        *webrtc.SessionDescription
		Active       bool
		RtpSessionId uuid.UUID
	}, error)

	ListenLobby(ctx context.Context, liveStreamId uuid.UUID, userId uuid.UUID, offer *webrtc.SessionDescription) (struct {
		Answer       *webrtc.SessionDescription
		Active       bool
		RtpSessionId uuid.UUID
	}, error)

	LeaveLobby(ctx context.Context, liveStreamId uuid.UUID, userId uuid.UUID) (bool, error)
}

type Space struct {
	Id        string        `json:"Id" gorm:"primaryKey"`
	ChannelId uint          `json:"-" gorm:"not null"`
	Channel   *models.Actor `json:"-" gorm:"foreignKey:ChannelId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	AccountId uint          `json:"-" gorm:"not null;"`
	Account   *auth.Account `json:"-" gorm:"foreignKey:AccountId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	// @TODO Space should not have this properties. Redactor this. This make a lot of trouble
	LiveStreamRepo *LiveStreamRepository `gorm:"-"`
	lobby          lobbyListenAccessor   `gorm:"-"`
	store          storage               `gorm:"-"`
}

func newSpace(id string, lobby lobbyListenAccessor, store storage, repo *LiveStreamRepository) (*Space, error) {
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

func (s *Space) StartListenLobby(ctx context.Context, stream *LiveStream, id uuid.UUID) (*webrtc.SessionDescription, error) {
	resourceData, err := s.lobby.StartListenLobby(ctx, stream.UUID, id)
	if err != nil {
		return nil, fmt.Errorf("start listening lobby: %w", err)
	}
	if !resourceData.Active {
		return nil, ErrLobbyNotActive
	}
	return resourceData.Offer, nil
}

func (s *Space) ListenLobby(ctx context.Context, offer *webrtc.SessionDescription, stream *LiveStream, id uuid.UUID) (bool, error) {
	resourceData, err := s.lobby.ListenLobby(ctx, stream.UUID, id, offer)
	if err != nil {
		return false, fmt.Errorf("listening lobby: %w", err)
	}
	if !resourceData.Active {
		return false, ErrLobbyNotActive
	}
	return resourceData.Active, nil
}

func (s *Space) LeaveLobby(ctx context.Context, stream *LiveStream, userId uuid.UUID) (bool, error) {
	left, err := s.lobby.LeaveLobby(ctx, stream.UUID, userId)
	if err != nil {
		return false, fmt.Errorf("leave lobby: %w", err)
	}
	return left, nil
}
