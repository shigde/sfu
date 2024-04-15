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
	"github.com/shigde/sfu/internal/telemetry"
)

func offerEndpoint(ctx context.Context, sessionCxt context.Context, e *Engine, sessionId uuid.UUID, liveStream uuid.UUID, endpointType EndpointType, options ...EndpointOption) (*Endpoint, error) {
	_, span := newTraceSpan(ctx, sessionCxt, "rtp: offer_endpoint")
	defer span.End()
	metric.GraphNodeUpdate(metric.BuildNode(sessionId.String(), liveStream.String(), endpointType.ToString()))

	// bild rtp endpoint setup
	endpoint := newEndpoint(sessionCxt, sessionId.String(), liveStream.String(), endpointType, options...)

	// special setup for ingress
	if endpointType == IngressEndpoint {
		// check in the options was a dispatcher
		if endpoint.dispatcher == nil {
			return nil, telemetry.RecordErrorf(span, "setup ingress endpoint", errors.New("no track dispatcher found"))
		}
		endpoint.receiver = newReceiver(sessionCxt, sessionId, liveStream, endpoint.dispatcher, endpoint.trackSdpInfoRepository)
	}

	// Setup stats
	withStatsGetter := withOnStatsGetter(func(getter stats.Getter) {
		statsRegistry := rtpStats.NewRegistry(sessionId.String(), getter)
		// receiver only for ingress needed
		if endpoint.receiver != nil {
			endpoint.receiver.statsRegistry = statsRegistry
		}
		endpoint.statsRegistry = statsRegistry
	})

	api, err := e.createApi(withStatsGetter)
	if err != nil {
		return nil, telemetry.RecordErrorf(span, "creating api", err)
	}

	var pc *webrtc.PeerConnection
	if pc, err = api.NewPeerConnection(e.config); err != nil {
		return nil, telemetry.RecordErrorf(span, "create  peer connection", err)
	}
	endpoint.peerConnection = pc
	// receive tracks only needed for ingress
	if endpoint.receiver != nil {
		endpoint.peerConnection.OnTrack(endpoint.receiver.onTrack)
	}

	// sending tracks only for needed egress
	if endpointType == EgressEndpoint {
		setupOnNegotiationNeeded(sessionCxt, endpoint, sessionId, liveStream)
	}

	endpoint.peerConnection.OnICEConnectionStateChange(endpoint.onICEConnectionStateChange)

	if endpoint.onChannel != nil {
		err = creatDC(pc, endpoint.onChannel)
		if err != nil {
			return nil, fmt.Errorf("creating data channel: %w", err)
		}
	}

	offer, err := endpoint.peerConnection.CreateOffer(nil)
	if err != nil {
		return nil, fmt.Errorf("creating offer: %w", err)
	}

	endpoint.gatherComplete = webrtc.GatheringCompletePromise(pc)
	if err = pc.SetLocalDescription(offer); err != nil {
		return nil, err
	}
	return endpoint, nil
}
