package rtp

import "context"

type TrackDispatcher interface {
	DispatchAddTrack(ctx context.Context, track *TrackInfo)
	DispatchRemoveTrack(ctx context.Context, track *TrackInfo)
}
