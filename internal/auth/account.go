package auth

import (
	"github.com/shigde/sfu/internal/activitypub/models"
	"gorm.io/gorm"
)

type Account struct {
	User     string        `gorm:"index;unique"`
	Email    string        `gorm:"index;unique"`
	UUID     string        `gorm:"index;unique"`
	Password string        `gorm:"not null"`
	Active   bool          `gorm:"not null,default:false"`
	ActorId  uint          `gorm:"not null;unique"`
	Actor    *models.Actor `gorm:"foreignKey:ActorId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	gorm.Model
}
