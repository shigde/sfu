package rtp

import (
	"github.com/pion/webrtc/v3"
)

type EndpointOption func(*Endpoint)

func EndpointWithOnEstablished(onEstablished func()) func(endpoint *Endpoint) {
	return func(endpoint *Endpoint) {
		endpoint.onEstablished = onEstablished
	}
}

func EndpointWithTrack(track webrtc.TrackLocal) func(endpoint *Endpoint) {
	return func(endpoint *Endpoint) {
		endpoint.AddTrack(track)
	}
}
