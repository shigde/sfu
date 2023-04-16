package media

import (
	"sync"

	"github.com/google/uuid"
)

type StreamRepository struct {
	locker  *sync.RWMutex
	streams []StreamResource
}

func newStreamRepository() *StreamRepository {
	streams := []StreamResource{
		{Id: uuid.New().String()},
		{Id: uuid.New().String()},
	}
	return &StreamRepository{
		&sync.RWMutex{},
		streams,
	}
}

func (r *StreamRepository) AddStream(s StreamResource) string {
	r.locker.Lock()
	defer r.locker.Unlock()
	s.Id = uuid.New().String()
	r.streams = append(r.streams, s)
	return s.Id
}

func (r *StreamRepository) AllStreams() []StreamResource {
	r.locker.RLock()
	defer r.locker.RUnlock()
	return r.streams
}

func (r *StreamRepository) StreamById(id string) (StreamResource, bool) {
	r.locker.RLock()
	defer r.locker.RUnlock()
	for _, stream := range r.streams {
		if stream.Id == id {
			return stream, true
		}
	}
	return StreamResource{}, false
}

func (r *StreamRepository) DeleteStream(id string) bool {
	r.locker.Lock()
	defer r.locker.Unlock()
	if i := index(id, r.streams); i != -1 {
		r.streams = append(r.streams[:i], r.streams[i+1:]...)
		return true
	}
	return false
}

func index(id string, resources []StreamResource) int {
	for i, stream := range resources {
		if stream.Id == id {
			return i
		}
	}
	return -1
}

func (r *StreamRepository) Contains(id string) bool {
	r.locker.RLock()
	defer r.locker.RUnlock()
	for _, stream := range r.AllStreams() {
		if stream.Id == id {
			return true
		}
	}
	return false
}

func (r *StreamRepository) StreamUpdate(stream StreamResource) bool {
	r.locker.Lock()
	defer r.locker.Unlock()
	if i := index(stream.Id, r.streams); i != -1 {
		r.streams[i] = stream
		return true
	}
	return false
}
