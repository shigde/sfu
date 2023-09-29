package models

import (
	"gorm.io/gorm"
)

type ActorFollowState int

const (
	Accepted ActorFollowState = iota
	Pending
	Rejected
)

func (fs ActorFollowState) String() string {
	return []string{"accepted", "pending", "rejected"}[fs]
}

type ActorFollow struct {
	ActorId       uint   `gorm:"actorId;not null"`
	TargetActorId uint   `gorm:"targetActorId;not null"`
	State         string `gorm:"state;not null"`
	Score         uint   `gorm:"score;default:20;not null"`
	Iri           string `gorm:"url;not null"`
	gorm.Model
}
