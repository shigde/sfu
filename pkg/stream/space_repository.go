package stream

import (
	"sync"
)

type SpaceRepository struct {
	locker *sync.RWMutex
	space  map[string]*Space
	lobby  LobbyJoiner
}

func newSpaceRepository(lobby LobbyJoiner) *SpaceRepository {
	space := make(map[string]*Space)
	return &SpaceRepository{
		&sync.RWMutex{},
		space,
		lobby,
	}
}

func (r *SpaceRepository) GetOrCreateSpace(id string) *Space {
	r.locker.Lock()
	defer r.locker.Unlock()
	currentSpace, ok := r.space[id]
	if !ok {
		space := newSpace(id, r.lobby)
		r.space[id] = space
		return space
	}
	return currentSpace
}

func (r *SpaceRepository) GetSpace(id string) (*Space, bool) {
	r.locker.Lock()
	defer r.locker.Unlock()
	currentSpace, ok := r.space[id]
	return currentSpace, ok
}

func (r *SpaceRepository) Delete(id string) bool {
	r.locker.Lock()
	defer r.locker.Unlock()
	if _, ok := r.space[id]; ok {
		delete(r.space, id)
		return true
	}
	return false
}

func (r *SpaceRepository) Len() int {
	r.locker.RLock()
	defer r.locker.RUnlock()
	return len(r.space)
}
