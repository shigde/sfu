package lobby

import (
	"sync"

	"github.com/google/uuid"
)

type sessionRepository struct {
	locker   *sync.RWMutex
	sessions map[uuid.UUID]*session
}

func newSessionRepository() *sessionRepository {
	sessions := make(map[uuid.UUID]*session)
	return &sessionRepository{
		&sync.RWMutex{},
		sessions,
	}
}

func (r *sessionRepository) Add(s *session) {
	r.locker.Lock()
	defer r.locker.Unlock()
	r.sessions[s.Id] = s
}

func (r *sessionRepository) All() map[uuid.UUID]*session {
	r.locker.RLock()
	defer r.locker.RUnlock()
	return r.sessions
}

func (r *sessionRepository) FindById(id uuid.UUID) (*session, bool) {
	r.locker.RLock()
	defer r.locker.RUnlock()
	session, ok := r.sessions[id]
	return session, ok
}

func (r *sessionRepository) FindByUserId(userId uuid.UUID) (*session, bool) {
	r.locker.RLock()
	defer r.locker.RUnlock()

	for _, session := range r.sessions {
		if session.user == userId {
			return session, true
		}
	}
	return nil, false
}

func (r *sessionRepository) Delete(id uuid.UUID) bool {
	r.locker.Lock()
	defer r.locker.Unlock()

	if _, ok := r.sessions[id]; ok {
		delete(r.sessions, id)
		return true
	}
	return false
}

func (r *sessionRepository) Contains(id uuid.UUID) bool {
	r.locker.RLock()
	defer r.locker.RUnlock()

	_, ok := r.sessions[id]
	return ok
}

func (r *sessionRepository) Update(s *session) bool {
	r.locker.Lock()
	defer r.locker.Unlock()
	if _, ok := r.sessions[s.Id]; ok {
		r.sessions[s.Id] = s
		return true
	}
	return false
}

func (r *sessionRepository) Len() int {
	r.locker.RLock()
	defer r.locker.RUnlock()
	return len(r.sessions)
}
