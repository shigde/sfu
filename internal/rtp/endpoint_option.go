package rtp

import (
	"context"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

type EndpointOption func(*Endpoint)

func EndpointWithTrack(track webrtc.TrackLocal, purpose Purpose) func(endpoint *Endpoint) {
	return func(endpoint *Endpoint) {
		if endpoint.initTracks == nil {
			endpoint.initTracks = make([]*initTrack, 0)
		}
		endpoint.initTracks = append(endpoint.initTracks, &initTrack{purpose: purpose, track: track})
	}
}

func EndpointWithGetCurrentTrackCbk(ckk func(ctx context.Context, sessionId uuid.UUID) ([]*TrackInfo, error)) func(endpoint *Endpoint) {
	return func(endpoint *Endpoint) {
		if endpoint.getCurrentTracksCbk == nil {
			endpoint.getCurrentTracksCbk = ckk
		}
	}
}

func EndpointWithOnEstablishedListener(onEstablished func()) func(endpoint *Endpoint) {
	return func(endpoint *Endpoint) {
		endpoint.onEstablished = onEstablished
	}
}

func EndpointWithNegotiationNeededListener(f func(sdp webrtc.SessionDescription)) func(endpoint *Endpoint) {
	return func(endpoint *Endpoint) {
		endpoint.onNegotiationNeeded = f
	}
}

func EndpointWithONNSetupTrigger(waitBeforeONNSetup <-chan struct{}) func(endpoint *Endpoint) {
	return func(endpoint *Endpoint) {
		endpoint.waitBeforeONNSetup = waitBeforeONNSetup
	}
}

func EndpointWithLostConnectionListener(f func()) func(endpoint *Endpoint) {
	return func(endpoint *Endpoint) {
		endpoint.onLostConnection = f
	}
}

func EndpointWithonIceStateConnectedListener(onIceStateConnected func()) func(endpoint *Endpoint) {
	return func(endpoint *Endpoint) {
		endpoint.onIceStateConnected = onIceStateConnected
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
