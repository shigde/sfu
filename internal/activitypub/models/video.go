package models

import (
	"net/url"

	"gorm.io/gorm"
)

type Video struct {
	Iri     string   `gorm:"iri;not null;index;"`
	OwnerId uint     `gorm:"owner_id;not null;"`
	Owner   *Actor   `gorm:"foreignKey:OwnerId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Guests  []*Actor `gorm:"many2many:video_guests;"`
	iriUrl  *url.URL `gorm:"-"`
	gorm.Model
}

func (s *Video) GetVideoIri() *url.URL {
	iri, _ := url.Parse(s.Iri)
	return iri
}
