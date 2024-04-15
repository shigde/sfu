package rtp

import (
	"context"

	"github.com/google/uuid"
	"golang.org/x/exp/slog"
)

func setupOnNegotiationNeeded(sessionCxt context.Context, endpoint *Endpoint, sessionId uuid.UUID, liveStream uuid.UUID) {

	// @TODO: Fix the race
	// First we create the egress endpoint and after this we add the individual tracks.
	// I don't know why, but Pion doesn't trigger renegotiation when creating a peer connection with tracks and the sdp
	// exchange is not finish. A peer connection without tracks where all tracks are added afterwards triggers renegotiation.
	// Unfortunately, "sendingTracks" could be outdated in the meantime.
	// This creates a race between remove and add track that I still have to think about it.
	if endpoint.getCurrentTracksCbk != nil {
		go func() {
			slog.Debug("rtp.establish_endpoint: add tracks", "sessionId", sessionId, "liveStream", liveStream)
			select {
			case <-sessionCxt.Done():
				return
			case <-endpoint.initComplete:
				// this is a data race with the lobby hub
				ctx, span := newTraceSpan(context.Background(), endpoint.sessionCxt, "endpoint_negotiation_add_track")
				if tracksList, err := endpoint.getCurrentTracksCbk(ctx, sessionId); err == nil {
					for _, trackInfo := range tracksList {
						endpoint.AddTrack(ctx, trackInfo)
					}
				}
				span.End()
			}
		}()
	}

	if endpoint.onNegotiationNeeded != nil {
		slog.Debug("rtp.engine: sender: OnNegotiationNeeded setup start")
		endpoint.peerConnection.OnNegotiationNeeded(endpoint.doRenegotiation)
		slog.Debug("rtp.engine: sender: OnNegotiationNeeded setup finish")
	}
}
