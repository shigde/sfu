package stream

import (
	"context"
	"fmt"
)

type SpaceManager struct {
	spaces *SpaceRepository
	lobby  lobbyListenAccessor
}

func NewSpaceManager(lobby lobbyListenAccessor, store storage) (*SpaceManager, error) {
	spaces, err := newSpaceRepository(lobby, store)
	if err != nil {
		return nil, fmt.Errorf("creating space repository: %w", err)
	}
	return &SpaceManager{spaces, lobby}, nil
}

func (m *SpaceManager) GetSpace(ctx context.Context, id string) (*Space, error) {
	return m.spaces.GetSpace(ctx, id)
}

func (m *SpaceManager) GetOrCreateSpace(ctx context.Context, id string) (*Space, error) {
	return m.spaces.GetOrCreateSpace(ctx, id)
}
