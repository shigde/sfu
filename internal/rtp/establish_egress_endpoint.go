package rtp

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/interceptor/pkg/stats"
	"github.com/pion/webrtc/v3"
	rtpStats "github.com/shigde/sfu/internal/rtp/stats"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/slog"
)

func EstablishEgressEndpoint(ctx context.Context, e *Engine, sessionId uuid.UUID, sendingTracks []webrtc.TrackLocal, handler StateEventHandler) (*Endpoint, error) {
	_, span := otel.Tracer(tracerName).Start(ctx, "engine:create ingress endpoint")
	defer span.End()

	endpoint := &Endpoint{}
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

	peerConnection.OnICEConnectionStateChange(handler.OnConnectionStateChange)

	initComplete := make(chan struct{})

	// @TODO: Fix the race
	// First we create the sender endpoint and after this we add the individual tracks.
	// I don't know why, but Pion doesn't trigger renegotiation when creating a peer connection with tracks and the sdp
	// exchange is not finish. A peer connection without tracks where all tracks are added afterwards triggers renegotiation.
	// Unfortunately, "sendingTracks" could be outdated in the meantime.
	// This creates a race between remove and add track that I still have to think about it.
	go func() {
		<-initComplete
		if sendingTracks != nil {
			for _, track := range sendingTracks {

				//peerConnection.GetTransceivers()[1].Sender().GetParameters().Codecs.Encodings[1].SSRC
				if _, err = peerConnection.AddTrack(track); err != nil {
					//a, b := s.GetParameters().Encodings[0].SSRC

					slog.Error("rtp.engine: adding track to connection", "err", err)
				}
			}
		}
	}()

	peerConnection.OnNegotiationNeeded(func() {
		<-initComplete
		slog.Debug("rtp.engine: sender OnNegotiationNeeded was triggered")
		offer, err := peerConnection.CreateOffer(nil)
		if err != nil {
			slog.Error("rtp.engine: sender OnNegotiationNeeded", "err", err)
			return
		}
		gg := webrtc.GatheringCompletePromise(peerConnection)
		_ = peerConnection.SetLocalDescription(offer)
		<-gg
		handler.OnNegotiationNeeded(*peerConnection.LocalDescription())
	})
	slog.Debug("rtp.engine: sender: OnNegotiationNeeded setup finish")

	err = creatDC(peerConnection, handler)
	if err != nil {
		return nil, fmt.Errorf("creating data channel: %w", err)
	}

	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		return nil, fmt.Errorf("creating offer: %w", err)
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	if err = peerConnection.SetLocalDescription(offer); err != nil {
		return nil, err
	}
	endpoint.peerConnection = peerConnection
	endpoint.gatherComplete = gatherComplete
	endpoint.initComplete = initComplete
	return endpoint, nil
}
