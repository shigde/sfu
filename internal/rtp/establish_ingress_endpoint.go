package rtp

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/interceptor/pkg/stats"
	"github.com/pion/webrtc/v3"
	rtpStats "github.com/shigde/sfu/internal/rtp/stats"
	"go.opentelemetry.io/otel"
)

func EstablishIngressEndpoint(ctx context.Context, e *Engine, sessionId uuid.UUID, offer webrtc.SessionDescription, options ...EndpointOption) (*Endpoint, error) {
	_, span := otel.Tracer(tracerName).Start(ctx, "engine:create egress endpoint")
	defer span.End()

	endpoint := &Endpoint{}
	for _, opt := range options {
		opt(endpoint)
	}

	trackInfos, err := getTrackInfo(offer, sessionId)
	if err != nil {
		return nil, fmt.Errorf("parsing track info: %w ", err)
	}

	if endpoint.dispatcher == nil {
		return nil, errors.New("no track dispatcher found")
	}

	receiver := newReceiver(sessionId, endpoint.dispatcher, trackInfos)
	withGetter := withOnStatsGetter(func(getter stats.Getter) {
		statsRegistry := rtpStats.NewRegistry(sessionId.String(), getter)
		receiver.statsRegistry = statsRegistry
		endpoint.statsRegistry = statsRegistry
	})

	api, err := e.createApi(withGetter)
	if err != nil {
		return nil, fmt.Errorf("creating api: %w", err)
	}

	peerConnection, err := api.NewPeerConnection(e.config)
	if err != nil {
		return nil, fmt.Errorf("create egress peer connection: %w ", err)
	}
	peerConnection.OnTrack(receiver.onTrack)

	if endpoint.onICEConnectionStateChange != nil {
		peerConnection.OnICEConnectionStateChange(endpoint.onICEConnectionStateChange)
	}
	if endpoint.onChannel != nil {
		peerConnection.OnDataChannel(endpoint.onChannel)
	}

	if err := peerConnection.SetRemoteDescription(offer); err != nil {
		return nil, err
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}

	if err = peerConnection.SetLocalDescription(answer); err != nil {
		return nil, err
	}

	endpoint.peerConnection = peerConnection
	endpoint.receiver = receiver
	endpoint.gatherComplete = gatherComplete
	return endpoint, nil
}
