package rtp

import (
	"errors"

	"github.com/pion/webrtc/v3"
	"golang.org/x/exp/slog"
)

var errSenderAlreadyClosed = errors.New("sender already closed")

type sender struct {
	conn            *webrtc.PeerConnection
	addTrackChan    <-chan *webrtc.TrackLocalStaticRTP
	removeTrackChan <-chan *webrtc.TrackLocalStaticRTP
	quit            chan struct{}
}

func newSender(conn *webrtc.PeerConnection) *sender {
	addTrack := make(<-chan *webrtc.TrackLocalStaticRTP, 1)
	removeTrack := make(<-chan *webrtc.TrackLocalStaticRTP, 1)
	quit := make(chan struct{})
	sender := &sender{
		conn,
		addTrack,
		removeTrack,
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
		case track := <-s.removeTrackChan:
			s.removeTrackFromConnection(track)
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

func (s *sender) removeTrackFromConnection(track *webrtc.TrackLocalStaticRTP) {
	senders := s.conn.GetSenders()
	for _, sender := range senders {
		if sender.Track().ID() == track.ID() {
			if err := s.conn.RemoveTrack(sender); err != nil {
				slog.Error("rtc.sender: remove track from connection",
					"stream", track.StreamID(),
					"track", track.ID(),
					"err", err,
				)
			}
		}
	}
}
