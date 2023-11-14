package lobby

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LobbyEntity struct {
	LiveStreamId uuid.UUID `json:"streamId"`
	UUID         uuid.UUID `json:"-"`
	IsRunning    bool      `json:"isLobbyRunning"`
	IsLive       bool      `json:"isLive"`
	gorm.Model
}

func NewLobbyEntity(streamID uuid.UUID) *LobbyEntity {
	return &LobbyEntity{
		LiveStreamId: streamID,
		UUID:         uuid.New(),
		IsRunning:    false,
		IsLive:       false,
	}
}

func (LobbyEntity) TableName() string {
	return "lobbies"
}
