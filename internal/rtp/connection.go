package rtp

import (
	"context"

	"github.com/pion/webrtc/v3"
)

type Connetcion struct {
	PeerConnection *webrtc.PeerConnection
	GatherComplete <-chan struct{}
}

func (c *Connetcion) GetLocalDescription(ctx context.Context) (*webrtc.SessionDescription, error) {
	select {
	case <-c.GatherComplete:
		return c.PeerConnection.LocalDescription(), nil
	case <-ctx.Done():
		return nil, ErrIceGatheringInteruption
	}
}
