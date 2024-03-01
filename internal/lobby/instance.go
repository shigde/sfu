package lobby

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Instance struct {
	UUID  uuid.UUID // my id for this instance like user id
	Token bool
	Host  string
	gorm.Model
}
