package rtp

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/interceptor/pkg/stats"
	"github.com/pion/webrtc/v3"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/slog"
)

func EstablishEgressEndpoint(ctx context.Context, e *Engine, sessionId uuid.UUID, offer webrtc.SessionDescription, dispatcher TrackDispatcher, handler StateEventHandler) (*Endpoint, error) {
	_, span := otel.Tracer(tracerName).Start(ctx, "engine:create egress endpoint")
	defer span.End()

	trackInfos, err := getTrackInfo(offer, sessionId)
	if err != nil {
		return nil, fmt.Errorf("parsing track info: %w ", err)
	}

	receiver := newReceiver(sessionId, dispatcher, trackInfos)
	withGetter := withOnStatsGetter(func(getter stats.Getter) {
		receiver.stats = getter
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

	peerConnection.OnICEConnectionStateChange(handler.OnConnectionStateChange)

	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		slog.Debug("rtp.engine: egress endpoint new DataChannel", "label", d.Label(), "id", d.ID())
		handler.OnChannel(d)
	})

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

	return &Endpoint{
		peerConnection: peerConnection,
		receiver:       receiver,
		gatherComplete: gatherComplete,
	}, nil
}
