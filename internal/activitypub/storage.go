package activitypub

import (
	"time"

	"gorm.io/gorm"
)

const queryTimeOut = 5 * time.Second

type Storage interface {
	GetDatabase() *gorm.DB
}

// gorm.Model definition
type entity struct {
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
