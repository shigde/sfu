package stream

type SpaceManager struct {
	spaces *SpaceRepository
	lobby  lobbyAccessor
}

func NewSpaceManager(lobby lobbyAccessor) *SpaceManager {
	spaces := newSpaceRepository(lobby)
	return &SpaceManager{spaces, lobby}
}

func (m *SpaceManager) GetSpace(id string) (*Space, bool) {
	return m.spaces.GetSpace(id)
}

func (m *SpaceManager) GetOrCreateSpace(id string) *Space {
	return m.spaces.GetOrCreateSpace(id)
}
