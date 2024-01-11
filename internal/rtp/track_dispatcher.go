package rtp

type TrackDispatcher interface {
	DispatchAddTrack(track *TrackInfo)
	DispatchRemoveTrack(track *TrackInfo)
}
