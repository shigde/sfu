package lobby

import (
	"net/url"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LobbyEntity struct {
	LiveStreamId uuid.UUID `json:"streamId"`
	UUID         uuid.UUID `json:"-"`
	IsRunning    bool      `json:"isLobbyRunning"`
	IsLive       bool      `json:"isLive"`
	Host         string    `json:"-"`
	gorm.Model
}

func NewLobbyEntity(streamID uuid.UUID) *LobbyEntity {
	// @TODO: The host should be send by Activity Pub
	host, _ := url.Parse("http://localhost:8080/federation/accounts/shig")
	return &LobbyEntity{
		LiveStreamId: streamID,
		UUID:         uuid.New(),
		IsRunning:    false,
		IsLive:       false,
		Host:         host.String(),
	}
}

func (LobbyEntity) TableName() string {
	return "lobbies"
}
