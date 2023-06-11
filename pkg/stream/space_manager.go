package stream

type SpaceManager struct {
	spaces *SpaceRepository
}

func NewSpaceManager() *SpaceManager {
	spaces := newSpaceRepository()
	return &SpaceManager{spaces}
}

func (m *SpaceManager) GetSpace(id string) (*Space, bool) {
	return m.spaces.GetSpace(id)
}

func (m *SpaceManager) GetOrCreateSpace(id string) *Space {
	return m.spaces.GetOrCreateSpace(id)
}
