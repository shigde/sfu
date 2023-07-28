package rtp

import (
	"context"
	"errors"

	"github.com/pion/webrtc/v3"
	"golang.org/x/exp/slog"
)

type Connection struct {
	peerConnection peerConnection
	receiver       *receiver
	sender         *sender
	AddTrackChan   <-chan *webrtc.TrackLocalStaticRTP
	gatherComplete <-chan struct{}
	closed         chan struct{}
}

func (c *Connection) GetAnswer(ctx context.Context) (*webrtc.SessionDescription, error) {
	// block until ice gathering is complete before return local sdp
	// all ice candidates should be part of the answer
	select {
	case <-c.gatherComplete:
		return c.peerConnection.LocalDescription(), nil
	case <-ctx.Done():
		return nil, errors.New("getting answer get interrupted")
	}
}

func (c *Connection) hasTrack(track *webrtc.TrackLocalStaticRTP) bool {
	slog.Debug("rtp.connection: has Tracks")
	rtpSenderList := c.peerConnection.GetSenders()
	for _, rtpSender := range rtpSenderList {
		if rtpTrack := rtpSender.Track(); rtpTrack != nil {
			if rtpTrack.ID() == track.ID() {
				return true
			}
		}
	}
	return false
}

func (c *Connection) AddTrack(track *webrtc.TrackLocalStaticRTP) bool {
	slog.Debug("rtp.connection: Add Track")
	if has := c.hasTrack(track); !has {
		_, err := c.peerConnection.AddTrack(track)
		slog.Debug("rtp.connection: Add Tracks to connection", "err", err)
	}
	return false
}

func (c *Connection) GetTracks() []*webrtc.TrackLocalStaticRTP {
	slog.Debug("rtp.connection: get Tracks")
	return c.receiver.getAllTracks()
}

type peerConnection interface {
	LocalDescription() *webrtc.SessionDescription
	GetSenders() (result []*webrtc.RTPSender)
	AddTrack(track webrtc.TrackLocal) (*webrtc.RTPSender, error)
}
