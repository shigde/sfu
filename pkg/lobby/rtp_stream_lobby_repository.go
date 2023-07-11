package lobby

import (
	"sync"

	"github.com/google/uuid"
)

type RtpStreamLobbyRepository struct {
	locker  *sync.RWMutex
	lobbies map[uuid.UUID]*RtpStreamLobby
}

func newRtpStreamLobbyRepository() *RtpStreamLobbyRepository {
	lobbies := make(map[uuid.UUID]*RtpStreamLobby)
	return &RtpStreamLobbyRepository{
		&sync.RWMutex{},
		lobbies,
	}
}

func (r *RtpStreamLobbyRepository) getOrCreateLobby(id uuid.UUID) *RtpStreamLobby {
	r.locker.Lock()
	defer r.locker.Unlock()
	currentLobby, ok := r.lobbies[id]
	if !ok {
		lobby := newRtpStreamLobby(id)
		r.lobbies[id] = lobby
		return lobby
	}
	return currentLobby
}

func (r *RtpStreamLobbyRepository) getLobby(id uuid.UUID) (*RtpStreamLobby, bool) {
	r.locker.Lock()
	defer r.locker.Unlock()
	currentLobby, ok := r.lobbies[id]
	return currentLobby, ok
}

func (r *RtpStreamLobbyRepository) Delete(id uuid.UUID) bool {
	r.locker.Lock()
	defer r.locker.Unlock()
	if _, ok := r.lobbies[id]; ok {
		delete(r.lobbies, id)
		return true
	}
	return false
}

func (r *RtpStreamLobbyRepository) Len() int {
	r.locker.RLock()
	defer r.locker.RUnlock()
	return len(r.lobbies)
}
