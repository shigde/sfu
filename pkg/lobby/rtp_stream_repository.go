package lobby

import (
	"sync"

	"github.com/google/uuid"
)

type rtpStreamRepository struct {
	locker  *sync.RWMutex
	streams []*RtpStream
}

func newRtpStreamRepository() *rtpStreamRepository {
	var streams []*RtpStream
	return &rtpStreamRepository{
		&sync.RWMutex{},
		streams,
	}
}

func (r *rtpStreamRepository) Add(s *RtpStream) string {
	r.locker.Lock()
	defer r.locker.Unlock()
	s.Id = uuid.New().String()
	r.streams = append(r.streams, s)
	return s.Id
}

func (r *rtpStreamRepository) All() []*RtpStream {
	r.locker.RLock()
	defer r.locker.RUnlock()
	return r.streams
}

func (r *rtpStreamRepository) FindById(id string) (*RtpStream, bool) {
	r.locker.RLock()
	defer r.locker.RUnlock()
	for _, stream := range r.streams {
		if stream.Id == id {
			return stream, true
		}
	}
	return nil, false
}

func (r *rtpStreamRepository) Delete(id string) bool {
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

func (r *rtpStreamRepository) Contains(id string) bool {
	r.locker.RLock()
	defer r.locker.RUnlock()
	for _, stream := range r.All() {
		if stream.Id == id {
			return true
		}
	}
	return false
}

func (r *rtpStreamRepository) Update(stream *RtpStream) bool {
	r.locker.Lock()
	defer r.locker.Unlock()
	if i := index(stream.Id, r.streams); i != -1 {
		r.streams[i] = stream
		return true
	}
	return false
}

func (r *rtpStreamRepository) Len() int {
	r.locker.RLock()
	defer r.locker.RUnlock()
	return len(r.streams)
}
