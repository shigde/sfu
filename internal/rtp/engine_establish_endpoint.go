package rtp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/pion/interceptor/pkg/stats"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/metric"
	rtpStats "github.com/shigde/sfu/internal/rtp/stats"
	"github.com/shigde/sfu/internal/telemetry"
)

func establishEndpoint(ctx context.Context, sessionCxt context.Context, e *Engine, sessionId uuid.UUID, liveStream uuid.UUID, offer webrtc.SessionDescription, endpointType EndpointType, options ...EndpointOption) (*Endpoint, error) {
	_, span := newTraceSpan(ctx, sessionCxt, "rtp: establish_endpoint")
	defer span.End()
	metric.GraphNodeUpdate(metric.BuildNode(sessionId.String(), liveStream.String(), endpointType.ToString()))

	// bild rtp endpoint setup
	endpoint := newEndpoint(sessionCxt, sessionId.String(), liveStream.String(), endpointType, options...)

	// special setup for ingress
	if endpointType == IngressEndpoint {
		if err := getIngressTrackSdpInfo(offer, sessionId, endpoint.trackSdpInfoRepository); err != nil {
			return nil, telemetry.RecordErrorf(span, "parsing track info", err)
		}
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

	if endpoint.peerConnection, err = api.NewPeerConnection(e.config); err != nil {
		return nil, telemetry.RecordErrorf(span, "create  peer connection", err)
	}

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
		endpoint.peerConnection.OnDataChannel(endpoint.onChannel)
	}

	if err := endpoint.peerConnection.SetRemoteDescription(offer); err != nil {
		return nil, telemetry.RecordErrorf(span, "setup offer", err)
	}

	endpoint.gatherComplete = webrtc.GatheringCompletePromise(endpoint.getPeerConnection())
	answer, err := endpoint.peerConnection.CreateAnswer(nil)
	if err != nil {
		return nil, telemetry.RecordErrorf(span, "create answer", err)
	}

	if err = endpoint.peerConnection.SetLocalDescription(answer); err != nil {
		return nil, telemetry.RecordErrorf(span, "setup answer", err)
	}
	endpoint.SetInitComplete()
	return endpoint, nil
}
