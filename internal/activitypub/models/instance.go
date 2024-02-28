package models

import (
	"gorm.io/gorm"
)

type Instance struct {
	ActorId uint   `gorm:"not null;"`
	Actor   *Actor `gorm:"foreignKey:ActorId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	gorm.Model
}

func NewInstance(actor *Actor) *Instance {
	return &Instance{
		ActorId: actor.ID,
		Actor:   actor,
	}
}
