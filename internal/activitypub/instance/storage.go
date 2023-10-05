package instance

import (
	"context"
	"time"

	"gorm.io/gorm"
)

const queryTimeOut = 5 * time.Second

type Storage interface {
	GetDatabase() *gorm.DB
	GetDatabaseWithContext(ctx context.Context) (*gorm.DB, context.CancelFunc)
}
