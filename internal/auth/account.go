package auth

import (
	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/activitypub/models"
	"gorm.io/gorm"
)

type Account struct {
	UUID       uuid.UUID     `gorm:"index;unique"`
	Identifier string        `gorm:"index;unique"`
	ActorId    uint          `gorm:"not null"`
	Actor      *models.Actor `gorm:"foreignKey:ActorId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	gorm.Model
}
