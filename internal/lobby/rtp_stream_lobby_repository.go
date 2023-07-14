package lobby

import (
	"sync"

	"github.com/google/uuid"
)

type RtpStreamLobbyRepository struct {
	locker    *sync.RWMutex
	lobbies   map[uuid.UUID]*rtpStreamLobby
	rtpEngine rtpEngine
}

func newRtpStreamLobbyRepository(rtpEngine rtpEngine) *RtpStreamLobbyRepository {
	lobbies := make(map[uuid.UUID]*rtpStreamLobby)
	return &RtpStreamLobbyRepository{
		&sync.RWMutex{},
		lobbies,
		rtpEngine,
	}
}

func (r *RtpStreamLobbyRepository) getOrCreateLobby(id uuid.UUID) *rtpStreamLobby {
	r.locker.Lock()
	defer r.locker.Unlock()
	currentLobby, ok := r.lobbies[id]
	if !ok {
		lobby := newRtpStreamLobby(id, r.rtpEngine)
		r.lobbies[id] = lobby
		return lobby
	}
	return currentLobby
}

func (r *RtpStreamLobbyRepository) getLobby(id uuid.UUID) (*rtpStreamLobby, bool) {
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
