package lobby

import (
	"sync"

	"github.com/google/uuid"
)

type lobbyRepository struct {
	locker    *sync.RWMutex
	lobbies   map[uuid.UUID]*lobby
	rtpEngine rtpEngine
}

func newLobbyRepository(rtpEngine rtpEngine) *lobbyRepository {
	lobbies := make(map[uuid.UUID]*lobby)
	return &lobbyRepository{
		&sync.RWMutex{},
		lobbies,
		rtpEngine,
	}
}

func (r *lobbyRepository) getOrCreateLobby(id uuid.UUID) *lobby {
	r.locker.Lock()
	defer r.locker.Unlock()
	currentLobby, ok := r.lobbies[id]
	if !ok {
		lobby := newLobby(id, r.rtpEngine)
		r.lobbies[id] = lobby
		return lobby
	}
	return currentLobby
}

func (r *lobbyRepository) getLobby(id uuid.UUID) (*lobby, bool) {
	r.locker.Lock()
	defer r.locker.Unlock()
	currentLobby, ok := r.lobbies[id]
	return currentLobby, ok
}

func (r *lobbyRepository) Delete(id uuid.UUID) bool {
	r.locker.Lock()
	defer r.locker.Unlock()
	if _, ok := r.lobbies[id]; ok {
		delete(r.lobbies, id)
		return true
	}
	return false
}

func (r *lobbyRepository) Len() int {
	r.locker.RLock()
	defer r.locker.RUnlock()
	return len(r.lobbies)
}
