package stream

import (
	"context"
)

type SpaceManager struct {
	spaces *SpaceRepository
	lobby  lobbyListenAccessor
}

func NewSpaceManager(lobby lobbyListenAccessor, store storage, liveRepo *LiveStreamRepository) *SpaceManager {
	spaces := newSpaceRepository(lobby, store, liveRepo)
	return &SpaceManager{spaces, lobby}
}

func (m *SpaceManager) GetSpace(ctx context.Context, id string) (*Space, error) {
	return m.spaces.GetSpace(ctx, id)
}

func (m *SpaceManager) GetOrCreateSpace(ctx context.Context, id string) (*Space, error) {
	return m.spaces.GetOrCreateSpace(ctx, id)
}
