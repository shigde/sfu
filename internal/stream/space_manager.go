package stream

import (
	"context"
)

type SpaceManager struct {
	spaces *SpaceRepository
}

func NewSpaceManager(store storage) *SpaceManager {
	spaces := NewSpaceRepository(store)
	return &SpaceManager{spaces}
}

func (m *SpaceManager) GetSpace(ctx context.Context, identifier string) (*Space, error) {
	return m.spaces.GetByIdentifier(ctx, identifier)
}
