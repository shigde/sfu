package rtp

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/interceptor/pkg/stats"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/metric"
	rtpStats "github.com/shigde/sfu/internal/rtp/stats"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/slog"
)

func EstablishEgressEndpoint(sessionCxt context.Context, e *Engine, sessionId uuid.UUID, liveStream uuid.UUID, options ...EndpointOption) (*Endpoint, error) {
	metric.GraphNodeUpdate(metric.BuildNode(sessionId.String(), liveStream.String(), "egress"))
	_, span := otel.Tracer(tracerName).Start(sessionCxt, "rtp:establish_egress_endpoint")
	defer span.End()

	endpoint := newEndpoint(sessionCxt, sessionId.String(), liveStream.String(), EgressEndpoint, options...)
	withStatsGetter := withOnStatsGetter(func(getter stats.Getter) {
		endpoint.statsRegistry = rtpStats.NewRegistry(sessionId.String(), getter)
	})

	api, err := e.createApi(withStatsGetter)
	if err != nil {
		return nil, fmt.Errorf("creating api: %w", err)
	}

	peerConnection, err := api.NewPeerConnection(e.config)
	if err != nil {
		return nil, fmt.Errorf("create sender peer connection: %w ", err)
	}
	endpoint.peerConnection = peerConnection
	endpoint.peerConnection.OnICEConnectionStateChange(endpoint.onICEConnectionStateChange)

	initComplete := make(chan struct{})

	// @TODO: Fix the race
	// First we create the egress endpoint and after this we add the individual tracks.
	// I don't know why, but Pion doesn't trigger renegotiation when creating a peer connection with tracks and the sdp
	// exchange is not finish. A peer connection without tracks where all tracks are added afterwards triggers renegotiation.
	// Unfortunately, "sendingTracks" could be outdated in the meantime.
	// This creates a race between remove and add track that I still have to think about it.
	if endpoint.getCurrentTracksCbk != nil {
		go func() {
			slog.Debug("rtp.establish_egress: add tracks", "sessionId", sessionId, "liveStream", liveStream)
			select {
			case <-sessionCxt.Done():
				return
			case <-initComplete:
				if tracksList, err := endpoint.getCurrentTracksCbk(sessionId); err == nil {
					for _, trackInfo := range tracksList {
						endpoint.AddTrack(trackInfo)
					}
				}
			}
		}()
	}
	endpoint.initComplete = initComplete
	if endpoint.onNegotiationNeeded != nil {
		peerConnection.OnNegotiationNeeded(endpoint.doRenegotiation)
		slog.Debug("rtp.engine: sender: OnNegotiationNeeded setup finish")
	}

	if endpoint.onChannel != nil {
		err = creatDC(peerConnection, endpoint.onChannel)
		if err != nil {
			return nil, fmt.Errorf("creating data channel: %w", err)
		}
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
