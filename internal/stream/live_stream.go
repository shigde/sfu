package stream

import (
	"time"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/auth"
	"github.com/shigde/sfu/internal/lobby"
	"gorm.io/gorm"
)

type LiveStream struct {
	VideoId   string             `json:"-" gorm:"not null;"`
	Video     *models.Video      `json:"-" gorm:"foreignKey:VideoId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	UUID      uuid.UUID          `json:"uuid"`
	Title     string             `json:"title" gorm:"-"`
	LobbyId   uint               `json:"-" gorm:"not null;"`
	Lobby     *lobby.LobbyEntity `json:"-" gorm:"foreignKey:LobbyId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	SpaceId   uint               `json:"-" gorm:"not null;"`
	Space     *Space             `json:"-" gorm:"foreignKey:SpaceId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	AccountId uint               `json:"-" gorm:"not null;"`
	Account   *auth.Account      `json:"-" gorm:"foreignKey:AccountId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	User      string             `json:"user"`
	ID        uint               `json:"-" gorm:"primaryKey"`
	CreatedAt time.Time          `json:"-"`
	UpdatedAt time.Time          `json:"-"`
	DeletedAt gorm.DeletedAt     `json:"-" gorm:"index"`
}

func NewLiveStream(account *auth.Account, lobbyEntity *lobby.LobbyEntity, space *Space, video *models.Video) *LiveStream {
	streamID, _ := uuid.Parse(video.Uuid)
	stream := &LiveStream{}
	stream.Lobby = lobbyEntity
	stream.Account = account
	stream.Space = space
	stream.UUID = streamID
	stream.Video = video
	stream.User = account.User
	return stream

}
