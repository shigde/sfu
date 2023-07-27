package rtp

import (
	"context"
	"errors"

	"github.com/pion/webrtc/v3"
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
	if has := c.hasTrack(track); !has {
		_, _ = c.peerConnection.AddTrack(track)
	}
	return false
}

func (c *Connection) GetTracks() []*webrtc.TrackLocalStaticRTP {
	rtpSenderList := c.peerConnection.GetSenders()
	var tracks []*webrtc.TrackLocalStaticRTP
	for _, rtpSender := range rtpSenderList {
		if rtpTrack := rtpSender.Track(); rtpTrack != nil {
			localTrack, _ := rtpTrack.(*webrtc.TrackLocalStaticRTP)
			tracks = append(tracks, localTrack)
		}
	}
	return tracks
}

type peerConnection interface {
	LocalDescription() *webrtc.SessionDescription
	GetSenders() (result []*webrtc.RTPSender)
	AddTrack(track webrtc.TrackLocal) (*webrtc.RTPSender, error)
}
