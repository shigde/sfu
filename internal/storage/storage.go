package storage

import (
	"context"

	"gorm.io/gorm"
)

type Storage interface {
	GetDatabase() *gorm.DB
	GetDatabaseWithContext(ctx context.Context) (*gorm.DB, context.CancelFunc)
}
