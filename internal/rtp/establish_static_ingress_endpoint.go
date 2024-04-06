package rtp

import (
	"context"
	"fmt"

	"github.com/pion/webrtc/v3"
	"golang.org/x/exp/slog"
)

// EstablishStaticIngressEndpoint
// This is used from cmd line toll to start a webrtc connection in a running lobby
// Deprecated: Because the Endpoint API is getting simpler
func EstablishStaticIngressEndpoint(ctx context.Context, e *Engine, sendingTracks []webrtc.TrackLocal, options ...EndpointOption) (*Endpoint, error) {
	stateHandler := newMediaStateEventHandler()
	api, err := e.createApi()
	if err != nil {
		return nil, fmt.Errorf("creating api: %w", err)
	}
	peerConnection, err := api.NewPeerConnection(e.config)
	if err != nil {
		return nil, fmt.Errorf("create receiver peer connection: %w ", err)
	}
	endpoint := &Endpoint{
		endpointType:   IngressEndpoint,
		peerConnection: peerConnection,
	}
	for _, opt := range options {
		opt(endpoint)
	}

	// This makes no sense, please find another way to deal with context
	_, iceConnectedCtxCancel := context.WithCancel(ctx)

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		slog.Debug("rtp.engine: connection State has changed", "state", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateConnected {
			iceConnectedCtxCancel()
		}
	})

	peerConnection.OnSignalingStateChange(func(signalState webrtc.SignalingState) {
		if signalState == webrtc.SignalingStateStable {
			if endpoint.onEstablished != nil {
				endpoint.onEstablished()
			}
		}
	})

	for _, track := range sendingTracks {
		if _, err := peerConnection.AddTrack(track); err != nil {
			return nil, err
		}
	}

	err = creatDC(peerConnection, stateHandler.OnChannel)

	if err != nil {
		return nil, fmt.Errorf("creating data channel: %w", err)
	}

	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		return nil, fmt.Errorf("creating offer: %w", err)
	}

	endpoint.gatherComplete = webrtc.GatheringCompletePromise(peerConnection)

	if err = peerConnection.SetLocalDescription(offer); err != nil {
		return nil, err
	}

	return endpoint, nil
}
