package engine

import (
	"fmt"

	"github.com/pion/webrtc/v3"
)

type Connection struct {
	peerConnection *webrtc.PeerConnection
}

func NewConnection() (*Connection, error) {
	api := webrtc.NewAPI()
	peerConnection, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return nil, fmt.Errorf("creating peer connection: %w", err)
	}

	return &Connection{
		peerConnection,
	}, nil
}
