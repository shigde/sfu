package rtp

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/interceptor/pkg/stats"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/metric"
	rtpStats "github.com/shigde/sfu/internal/rtp/stats"
	"go.opentelemetry.io/otel"
)

func EstablishIngressEndpoint(sessionCxt context.Context, e *Engine, sessionId uuid.UUID, liveStream uuid.UUID, offer webrtc.SessionDescription, options ...EndpointOption) (*Endpoint, error) {
	_, span := otel.Tracer(tracerName).Start(sessionCxt, "rtp:establish_ingress_endpoint")
	// Add node in dashboard
	metric.GraphNodeUpdate(metric.BuildNode(sessionId.String(), liveStream.String(), "ingress"))

	defer span.End()

	endpoint := newEndpoint(sessionCxt, sessionId.String(), liveStream.String(), IngressEndpoint, options...)

	err := getIngressTrackSdpInfo(offer, sessionId, endpoint.trackSdpInfoRepository)
	if err != nil {
		return nil, fmt.Errorf("parsing track info: %w ", err)
	}

	// check in the options was a dispatcher
	if endpoint.dispatcher == nil {
		return nil, errors.New("no track dispatcher found")
	}

	receiver := newReceiver(sessionCxt, sessionId, liveStream, endpoint.dispatcher, endpoint.trackSdpInfoRepository)
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

	peerConnection.OnICEConnectionStateChange(endpoint.onICEConnectionStateChange)

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
