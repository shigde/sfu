package stream

import (
	"time"

	"gorm.io/gorm"
)

const queryTimeOut = 5 * time.Second

type storage interface {
	GetDatabase() *gorm.DB
}

// gorm.Model definition
type entity struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
