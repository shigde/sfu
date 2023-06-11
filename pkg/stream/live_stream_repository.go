package stream

import (
	"sync"

	"github.com/google/uuid"
)

type LiveStreamRepository struct {
	locker  *sync.RWMutex
	streams []*LiveStream
}

func NewLiveStreamRepository() *LiveStreamRepository {
	var streams []*LiveStream
	return &LiveStreamRepository{
		&sync.RWMutex{},
		streams,
	}
}

func (r *LiveStreamRepository) Add(s *LiveStream) string {
	r.locker.Lock()
	defer r.locker.Unlock()
	s.Id = uuid.New().String()
	r.streams = append(r.streams, s)
	return s.Id
}

func (r *LiveStreamRepository) All() []*LiveStream {
	r.locker.RLock()
	defer r.locker.RUnlock()
	return r.streams
}

func (r *LiveStreamRepository) FindById(id string) (*LiveStream, bool) {
	r.locker.RLock()
	defer r.locker.RUnlock()
	for _, stream := range r.streams {
		if stream.Id == id {
			return stream, true
		}
	}
	return nil, false
}

func (r *LiveStreamRepository) Delete(id string) bool {
	r.locker.Lock()
	defer r.locker.Unlock()
	if i := index(id, r.streams); i != -1 {
		r.streams = append(r.streams[:i], r.streams[i+1:]...)
		return true
	}
	return false
}

func index(id string, resources []*LiveStream) int {
	for i, stream := range resources {
		if stream.Id == id {
			return i
		}
	}
	return -1
}

func (r *LiveStreamRepository) Contains(id string) bool {
	r.locker.RLock()
	defer r.locker.RUnlock()
	for _, stream := range r.All() {
		if stream.Id == id {
			return true
		}
	}
	return false
}

func (r *LiveStreamRepository) Update(stream *LiveStream) bool {
	r.locker.Lock()
	defer r.locker.Unlock()
	if i := index(stream.Id, r.streams); i != -1 {
		r.streams[i] = stream
		return true
	}
	return false
}

func (r *LiveStreamRepository) Len() int {
	r.locker.RLock()
	defer r.locker.RUnlock()
	return len(r.streams)
}
