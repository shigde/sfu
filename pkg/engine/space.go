package engine

type Space struct {
	Id               string `json:"Id"`
	publisher        []*Publisher
	PublicStreamRepo *RtpStreamRepository
}

func newSpace(id string) *Space {
	var publisher []*Publisher
	repo := NewRtpStreamRepository()
	return &Space{
		Id: id, publisher: publisher, PublicStreamRepo: repo,
	}
}

func (e *Space) Publish(Offer interface{}, user interface{}) (*interface{}, error) {
	return nil, nil
}
