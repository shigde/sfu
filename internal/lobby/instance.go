package lobby

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Instance struct {
	UUID  uuid.UUID
	Token bool
	Host  string
	gorm.Model
}
