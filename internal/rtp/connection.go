package rtp

import (
	"context"
	"errors"

	"github.com/pion/webrtc/v3"
)

type Connection struct {
	peerConnector  peerConnector
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
		return c.peerConnector.LocalDescription(), nil
	case <-ctx.Done():
		return nil, errors.New("getting answer get interrupted")
	}
}

type peerConnector interface {
	LocalDescription() *webrtc.SessionDescription
}

type ReceiverConnection struct {
	Connection
	receiver
}
