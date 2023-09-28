package models

import "gorm.io/gorm"

type Server struct {
	Host              string  `gorm:"host"`
	RedundancyAllowed bool    `gorm:"redundancyAllowed"`
	Actor             []Actor `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	gorm.Model
}
