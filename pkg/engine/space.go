package engine

type Space struct {
	Id               string `json:"Id"`
	publisher        []*Publisher
	publicStreamRepo *RtpStreamRepository
}

func newSpace(id string) *Space {
	var publisher []*Publisher
	repo := NewRtpStreamRepository()
	return &Space{
		Id: id, publisher: publisher, publicStreamRepo: repo,
	}
}
