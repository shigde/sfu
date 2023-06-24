package lobby

import (
	"sync"
)

type RtpStreamLobbyRepository struct {
	locker  *sync.RWMutex
	lobbies map[string]*RtpStreamLobby
}

func newRtpStreamLobbyRepository() *RtpStreamLobbyRepository {
	lobbies := make(map[string]*RtpStreamLobby)
	return &RtpStreamLobbyRepository{
		&sync.RWMutex{},
		lobbies,
	}
}

func (r *RtpStreamLobbyRepository) GetOrCreateLobby(id string) *RtpStreamLobby {
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

func (r *RtpStreamLobbyRepository) GetLobby(id string) (*RtpStreamLobby, bool) {
	r.locker.Lock()
	defer r.locker.Unlock()
	currentLobby, ok := r.lobbies[id]
	return currentLobby, ok
}

func (r *RtpStreamLobbyRepository) Delete(id string) bool {
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
