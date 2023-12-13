package rtp

import (
	"context"

	"github.com/pion/webrtc/v3"
)

type Connection struct {
	PeerConnection *webrtc.PeerConnection
	GatherComplete <-chan struct{}
}

func (c *Connection) GetLocalDescription(ctx context.Context) (*webrtc.SessionDescription, error) {
	select {
	case <-c.GatherComplete:
		return c.PeerConnection.LocalDescription(), nil
	case <-ctx.Done():
		return nil, ErrIceGatheringInterruption
	}
}
