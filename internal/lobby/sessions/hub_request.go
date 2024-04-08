package sessions

import (
	"context"

	"github.com/shigde/sfu/internal/rtp"
)

type hubRequest struct {
	ctx           context.Context
	kind          hubRequestKind
	track         *rtp.TrackInfo
	trackListChan chan<- []*rtp.TrackInfo
}

type hubRequestKind int

const (
	addTrack hubRequestKind = iota + 1
	removeTrack
	getTrackList
	muteTrack
)
