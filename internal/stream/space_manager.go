package stream

import (
	"context"
)

type SpaceManager struct {
	spaces *SpaceRepository
}

func NewSpaceManager(store storage, liveRepo *LiveStreamRepository) *SpaceManager {
	spaces := NewSpaceRepository(store)
	return &SpaceManager{spaces}
}

func (m *SpaceManager) GetSpace(ctx context.Context, id string) (*Space, error) {
	return m.spaces.GetSpace(ctx, id)
}
