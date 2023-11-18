package rtp

import (
	"context"
	"fmt"

	"github.com/pion/webrtc/v3"
	"golang.org/x/exp/slog"
)

func NewLocalStaticSenderEndpoint(e *Engine, sendingTracks []webrtc.TrackLocal, options ...EndpointOption) (*Endpoint, error) {
	stateHandler := newMediaStateEventHandler()
	peerConnection, err := e.api.NewPeerConnection(e.config)
	if err != nil {
		return nil, fmt.Errorf("create receiver peer connection: %w ", err)
	}
	endpoint := &Endpoint{peerConnection: peerConnection}
	for _, opt := range options {
		opt(endpoint)
	}

	// This makes no sense
	_, iceConnectedCtxCancel := context.WithCancel(context.Background())

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

	err = creatDC(peerConnection, stateHandler)

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
