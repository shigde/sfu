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

func NewLobbyEntity() *LobbyEntity {
	return &LobbyEntity{}
}

func (LobbyEntity) TableName() string {
	return "lobbies"
}
