package lobby

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LobbyEntity struct {
	LiveStreamId uuid.UUID
	UUID         uuid.UUID
	IsRunning    bool
	IsLive       bool
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
