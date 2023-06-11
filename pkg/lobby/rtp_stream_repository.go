package lobby

import (
	"sync"

	"github.com/google/uuid"
)

type RtpStreamRepository struct {
	locker  *sync.RWMutex
	streams []*RtpStream
}

func NewRtpStreamRepository() *RtpStreamRepository {
	var streams []*RtpStream
	return &RtpStreamRepository{
		&sync.RWMutex{},
		streams,
	}
}

func (r *RtpStreamRepository) Add(s *RtpStream) string {
	r.locker.Lock()
	defer r.locker.Unlock()
	s.Id = uuid.New().String()
	r.streams = append(r.streams, s)
	return s.Id
}

func (r *RtpStreamRepository) All() []*RtpStream {
	r.locker.RLock()
	defer r.locker.RUnlock()
	return r.streams
}

func (r *RtpStreamRepository) FindById(id string) (*RtpStream, bool) {
	r.locker.RLock()
	defer r.locker.RUnlock()
	for _, stream := range r.streams {
		if stream.Id == id {
			return stream, true
		}
	}
	return nil, false
}

func (r *RtpStreamRepository) Delete(id string) bool {
	r.locker.Lock()
	defer r.locker.Unlock()
	if i := index(id, r.streams); i != -1 {
		r.streams = append(r.streams[:i], r.streams[i+1:]...)
		return true
	}
	return false
}

func index(id string, resources []*RtpStream) int {
	for i, stream := range resources {
		if stream.Id == id {
			return i
		}
	}
	return -1
}

func (r *RtpStreamRepository) Contains(id string) bool {
	r.locker.RLock()
	defer r.locker.RUnlock()
	for _, stream := range r.All() {
		if stream.Id == id {
			return true
		}
	}
	return false
}

func (r *RtpStreamRepository) Update(stream *RtpStream) bool {
	r.locker.Lock()
	defer r.locker.Unlock()
	if i := index(stream.Id, r.streams); i != -1 {
		r.streams[i] = stream
		return true
	}
	return false
}

func (r *RtpStreamRepository) Len() int {
	r.locker.RLock()
	defer r.locker.RUnlock()
	return len(r.streams)
}
