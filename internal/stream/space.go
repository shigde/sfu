package stream

import (
	"errors"

	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/auth"
	"gorm.io/gorm"
)

var ErrLobbyNotActive = errors.New("lobby not active")

type Space struct {
	Identifier string        `gorm:"not null;unique;"`
	ChannelId  uint          `json:"-" gorm:"not null"`
	Channel    *models.Actor `json:"-" gorm:"foreignKey:ChannelId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	AccountId  uint          `json:"-" gorm:"not null;"`
	Account    *auth.Account `json:"-" gorm:"foreignKey:AccountId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	gorm.Model
}
