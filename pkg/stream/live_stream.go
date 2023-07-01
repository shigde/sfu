package stream

type LiveStream struct {
	Id      string `json:"Id"`
	SpaceId string `json:"-"`
	User    string `json:"-"`
}
