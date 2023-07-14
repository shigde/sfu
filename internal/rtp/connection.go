package rtp

import (
	"context"
	"errors"

	"github.com/pion/webrtc/v3"
)

type Connection struct {
	pc             *webrtc.PeerConnection
	receiver       *receiver
	sender         *sender
	gatherComplete <-chan struct{}
	closed         chan struct{}
}

func (c *Connection) GetAnswer(ctx context.Context) (*webrtc.SessionDescription, error) {
	// block until ice gathering is complete before return local sdp
	// all ice candidates should be part of the answer
	select {
	case <-c.gatherComplete:
		return c.pc.LocalDescription(), nil
	case <-ctx.Done():
		return nil, errors.New("getting answer get interrupted")
	}
}
