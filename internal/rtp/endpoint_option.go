package rtp

import (
	"github.com/pion/webrtc/v3"
	"golang.org/x/exp/slog"
)

type EndpointOption func(*Endpoint)

func EndpointWithOnEstablished(onEstablished func()) func(endpoint *Endpoint) {
	return func(endpoint *Endpoint) {
		endpoint.onEstablished = onEstablished
	}
}

func EndpointWithTrack(track webrtc.TrackLocal) func(endpoint *Endpoint) {
	return func(endpoint *Endpoint) {
		if _, err := endpoint.peerConnection.AddTrack(track); err != nil {
			slog.Error("add track", "err", err, "trackId", track.ID(), "streamId", track.StreamID())
		}
	}
}
