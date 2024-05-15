package lobby

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LobbyEntity struct {
	LiveStreamId uuid.UUID `json:"streamId" gorm:"not null;index;unique;"`
	UUID         uuid.UUID `json:"-"`
	Space        string
	IsRunning    bool   `json:"isLobbyRunning"`
	IsLive       bool   `json:"isLive"`
	Host         string `json:"-"`
	gorm.Model
}

func NewLobbyEntity(streamID uuid.UUID, space string, shigHost string) *LobbyEntity {

	return &LobbyEntity{
		UUID:         uuid.New(),
		IsRunning:    false,
		IsLive:       false,
		Space:        space,
		LiveStreamId: streamID,
		Host:         shigHost,
	}
}

func (LobbyEntity) TableName() string {
	return "lobbies"
}

func (e *LobbyEntity) GetHost() string {
	return e.Host
}
func (e *LobbyEntity) GetSpace() string {
	return e.Space
}
func (e *LobbyEntity) GetLiveStreamID() string {
	return e.LiveStreamId.String()
}
