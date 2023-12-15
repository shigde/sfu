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

func EndpointWithTrack(track webrtc.TrackLocal, purpose Purpose) func(endpoint *Endpoint) {
	return func(endpoint *Endpoint) {
		if endpoint.initTracks == nil {
			endpoint.initTracks = make([]*initTrack, 0)
		}
		endpoint.initTracks = append(endpoint.initTracks, &initTrack{purpose: purpose, track: track})
	}
}

func EndpointWithConnectionStateListener(f func(webrtc.ICEConnectionState)) func(endpoint *Endpoint) {
	return func(endpoint *Endpoint) {
		endpoint.onICEConnectionStateChange = f
	}
}

func EndpointWithNegotiationNeededListener(f func(sdp webrtc.SessionDescription)) func(endpoint *Endpoint) {
	return func(endpoint *Endpoint) {
		endpoint.onNegotiationNeeded = f
	}
}

func EndpointWithDataChannel(f func(dc *webrtc.DataChannel)) func(endpoint *Endpoint) {
	return func(endpoint *Endpoint) {
		endpoint.onChannel = f
	}
}

func EndpointWithTrackDispatcher(dispatcher TrackDispatcher) func(endpoint *Endpoint) {
	return func(endpoint *Endpoint) {
		endpoint.dispatcher = dispatcher
	}
}
