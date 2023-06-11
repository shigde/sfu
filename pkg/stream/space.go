package stream

type Space struct {
	Id             string `json:"Id"`
	LiveStreamRepo *LiveStreamRepository
}

func newSpace(id string) *Space {
	repo := NewLiveStreamRepository()
	return &Space{
		Id: id, LiveStreamRepo: repo,
	}
}

func (e *Space) Publish(Offer interface{}, user interface{}) (*interface{}, error) {
	return nil, nil
}
