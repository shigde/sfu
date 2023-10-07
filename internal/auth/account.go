package auth

import (
	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/activitypub/models"
	"gorm.io/gorm"
)

type Account struct {
	UUID       uuid.UUID `gorm:"index;unique"`
	Identifier string    `gorm:"index;unique"`
	Actor      models.Actor
	gorm.Model
}
