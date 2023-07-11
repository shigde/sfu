package stream

import "github.com/google/uuid"

type LiveStream struct {
	Id      string    `json:"Id" gorm:"primaryKey"`
	UUID    uuid.UUID `json:"-"`
	SpaceId string    `json:"-"`
	User    string    `json:"-"`
	entity
}
