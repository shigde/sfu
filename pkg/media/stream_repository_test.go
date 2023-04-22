package media

import "testing"

func TestAddStream(t *testing.T) {
	r := newStreamRepository()

	s := StreamResource{}
	r.AddStream(s)

	if len(r.AllStreams()) != 1 {
		t.Fatalf("wanted %d, got %d", 1, len(r.AllStreams()))
	}
}
