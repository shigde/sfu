package engine

import (
	"sync"
)

type SpaceRepository struct {
	locker *sync.RWMutex
	space  map[string]*Space
}

func newSpaceRepository() *RtpStreamRepository {
	var streams []*RtpStream
	return &RtpStreamRepository{
		&sync.RWMutex{},
		streams,
	}
}

func (r *SpaceRepository) GetOrCreateHour(id string) *Space {
	r.locker.Lock()
	defer r.locker.Unlock()
	currentSpace, ok := r.space[id]
	if !ok {
		return newSpace(id)
	}
	return currentSpace
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
