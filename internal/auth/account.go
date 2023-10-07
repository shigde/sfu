package auth

import (
	"github.com/shigde/sfu/internal/activitypub/models"
	"gorm.io/gorm"
)

type Account struct {
	User       string        `gorm:"index;unique"`
	Identifier string        `gorm:"index;unique"`
	ActorId    uint          `gorm:"not null;unique"`
	Actor      *models.Actor `gorm:"foreignKey:ActorId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	gorm.Model
}
