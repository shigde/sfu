package lobby

import (
	"net/url"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LobbyEntity struct {
	LiveStreamId uuid.UUID `json:"streamId"`
	UUID         uuid.UUID `json:"-"`
	Space        string
	IsRunning    bool   `json:"isLobbyRunning"`
	IsLive       bool   `json:"isLive"`
	Host         string `json:"-"`
	gorm.Model
}

func NewLobbyEntity(streamID uuid.UUID, space string) *LobbyEntity {
	// @TODO: The host should be send by Activity Pub
	host, _ := url.Parse("https://stream.shig.de/federation/accounts/shig")
	// host, _ := url.Parse("http://localhost:8080/federation/accounts/shig")
	return &LobbyEntity{
		UUID:         uuid.New(),
		IsRunning:    false,
		IsLive:       false,
		Space:        space,
		LiveStreamId: streamID,
		Host:         host.String(),
	}
}

func (LobbyEntity) TableName() string {
	return "lobbies"
}
