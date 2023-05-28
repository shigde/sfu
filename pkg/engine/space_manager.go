package engine

type SpaceManager struct {
}

func NewSpaceManager() *SpaceManager {
	return &SpaceManager{}
}

func (e *SpaceManager) Publish(Offer interface{}, user interface{}) (*interface{}, error) {
	return nil, nil
}
