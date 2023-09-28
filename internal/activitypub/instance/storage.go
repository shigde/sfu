package instance

import (
	"time"

	"gorm.io/gorm"
)

const queryTimeOut = 5 * time.Second

type Storage interface {
	GetDatabase() *gorm.DB
}
