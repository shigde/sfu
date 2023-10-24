package lobby

import "gorm.io/gorm"

type store interface {
	GetDatabase() *gorm.DB
}
