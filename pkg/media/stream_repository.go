package media

type StreamRepository struct {
	streams []StreamResource
}

func newStreamRepository() *StreamRepository {
	streams := []StreamResource{
		StreamResource{Id: "fsdfdgfshfsdghf"},
		StreamResource{Id: "Helhhcbdshcbhdsblo"},
	}
	return &StreamRepository{
		streams,
	}
}

func (r *StreamRepository) AddStream(s StreamResource) string {
	r.streams = append(r.streams, s)
	// @TODO create is
	return s.Id
}

func (r *StreamRepository) AllStreams() []StreamResource {
	return r.streams
}

func (r *StreamRepository) StreamById(id string) (StreamResource, bool) {
	for _, stream := range r.streams {
		if stream.Id == id {
			return stream, true
		}
	}
	return StreamResource{}, false
}

func (r *StreamRepository) DeleteStream(id string) bool {
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
	for _, stream := range r.AllStreams() {
		if stream.Id == id {
			return true
		}
	}
	return false
}

func (r *StreamRepository) StreamUpdate(stream StreamResource) bool {
	if i := index(stream.Id, r.streams); i != -1 {
		r.streams[i] = stream
		return true
	}
	return false
}
