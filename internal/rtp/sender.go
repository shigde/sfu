package rtp

import (
	"errors"
	"sync"

	"github.com/pion/webrtc/v3"
	"golang.org/x/exp/slog"
)

var errSenderAlreadyClosed = errors.New("sender already closed")

type sender struct {
	sync.RWMutex
	streams      map[string]*localStream
	conn         *webrtc.PeerConnection
	addTrackChan <-chan *webrtc.TrackLocalStaticRTP
	quit         chan struct{}
}

func newSender(conn *webrtc.PeerConnection) *sender {
	streams := make(map[string]*localStream)
	addTrack := make(<-chan *webrtc.TrackLocalStaticRTP)
	quit := make(chan struct{})
	sender := &sender{
		sync.RWMutex{},
		streams,
		conn,
		addTrack,
		quit,
	}

	go sender.run()
	return sender
}

func (s *sender) run() {
	for {
		select {
		case <-s.quit:
			return
		case track := <-s.addTrackChan:
			s.addTrackToConnection(track)
		}
	}
}

func (s *sender) stop() error {
	slog.Info("rtc.sender: stop")
	select {
	case <-s.quit:
		slog.Error("rtc.sender: the hub was already closed")
		return errSenderAlreadyClosed
	default:
		close(s.quit)
		slog.Info("rtc.sender: stopped was triggered")
	}
	return nil
}

func (s *sender) addTrackToConnection(track *webrtc.TrackLocalStaticRTP) {
	_, err := s.conn.AddTrack(track)
	if err != nil {
		slog.Error("rtc.sender: add track to connection",
			"stream", track.StreamID(),
			"track", track.ID(),
			"err", err,
		)
	}
}

func (r *sender) getStream(id string) *localStream {
	r.Unlock()
	defer r.Unlock()
	stream, ok := r.streams[id]
	if !ok {
		stream = newLocalStream(id)
		r.streams[id] = stream
	}
	return stream
}
