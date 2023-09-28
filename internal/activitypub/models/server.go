package models

import "gorm.io/gorm"

type Server struct {
	Host              string `gorm:"host"`
	RedundancyAllowed bool   `gorm:"redundancyAllowed"`
	gorm.Model
}
