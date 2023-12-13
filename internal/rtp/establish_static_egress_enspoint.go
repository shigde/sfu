package rtp

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"go.opentelemetry.io/otel"
)

func EstablishStaticEgressEndpoint(ctx context.Context, e *Engine, sessionId uuid.UUID, offer webrtc.SessionDescription, options ...EndpointOption) (*Endpoint, error) {
	_, span := otel.Tracer(tracerName).Start(ctx, "engine:create static egress endpoint")
	defer span.End()
	api, err := e.createApi()
	if err != nil {
		return nil, fmt.Errorf("creating api: %w", err)
	}

	peerConnection, err := api.NewPeerConnection(e.config)
	if err != nil {
		return nil, fmt.Errorf("create receiver peer connection: %w ", err)
	}
	endpoint := &Endpoint{
		peerConnection: peerConnection,
	}
	for _, opt := range options {
		opt(endpoint)
	}

	if err := peerConnection.SetRemoteDescription(offer); err != nil {
		return nil, err
	}

	endpoint.gatherComplete = webrtc.GatheringCompletePromise(peerConnection)
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}

	if err = peerConnection.SetLocalDescription(answer); err != nil {
		return nil, err
	}

	return endpoint, nil
}
