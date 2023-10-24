package stream

import (
	"time"

	"gorm.io/gorm"
)

const queryTimeOut = 5 * time.Second

type storage interface {
	GetDatabase() *gorm.DB
}
