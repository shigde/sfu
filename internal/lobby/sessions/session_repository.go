package sessions

import (
	"sync"

	"github.com/google/uuid"
)

type SessionRepository struct {
	locker   *sync.RWMutex
	sessions map[uuid.UUID]*Session
}

func NewSessionRepository() *SessionRepository {
	sessions := make(map[uuid.UUID]*Session)
	return &SessionRepository{
		&sync.RWMutex{},
		sessions,
	}
}

func (r *SessionRepository) New(s *Session) bool {
	r.locker.Lock()
	defer r.locker.Unlock()
	for _, session := range r.sessions {
		if session.user == s.user {
			return false
		}
	}
	r.sessions[s.Id] = s
	return true
}

func (r *SessionRepository) Add(s *Session) {
	r.locker.Lock()
	defer r.locker.Unlock()
	r.sessions[s.Id] = s
}

func (r *SessionRepository) All() map[uuid.UUID]*Session {
	r.locker.RLock()
	defer r.locker.RUnlock()
	return r.sessions
}

func (r *SessionRepository) FindById(id uuid.UUID) (*Session, bool) {
	r.locker.RLock()
	defer r.locker.RUnlock()
	session, ok := r.sessions[id]
	return session, ok
}

func (r *SessionRepository) FindByUserId(userId uuid.UUID) (*Session, bool) {
	r.locker.RLock()
	defer r.locker.RUnlock()

	for _, session := range r.sessions {
		if session.user == userId {
			return session, true
		}
	}
	return nil, false
}

func (r *SessionRepository) Delete(id uuid.UUID) bool {
	r.locker.Lock()
	defer r.locker.Unlock()

	if _, ok := r.sessions[id]; ok {
		delete(r.sessions, id)
		return true
	}
	return false
}

func (r *SessionRepository) Contains(id uuid.UUID) bool {
	r.locker.RLock()
	defer r.locker.RUnlock()

	_, ok := r.sessions[id]
	return ok
}

func (r *SessionRepository) Update(s *Session) bool {
	r.locker.Lock()
	defer r.locker.Unlock()
	if _, ok := r.sessions[s.Id]; ok {
		r.sessions[s.Id] = s
		return true
	}
	return false
}

func (r *SessionRepository) Len() int {
	r.locker.RLock()
	defer r.locker.RUnlock()
	return len(r.sessions)
}

func (r *SessionRepository) Iter(routine func(*Session)) {
	r.locker.Lock()
	defer r.locker.Unlock()

	for _, session := range r.sessions {
		routine(session)
	}
}

func (r *SessionRepository) DeleteByUser(user uuid.UUID) bool {
	r.locker.Lock()
	defer r.locker.Unlock()
	for _, session := range r.sessions {
		if session.user == user {
			delete(r.sessions, session.Id)
			return true
		}
	}

	return false
}
