package stream

type LiveStream struct {
	Id      string `json:"Id" gorm:"primaryKey"`
	SpaceId string `json:"-"`
	User    string `json:"-"`
	entity
}
