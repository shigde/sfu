package lobby

import "github.com/pion/webrtc/v3"

type hubRequest struct {
	kind          hubRequestKind
	track         *webrtc.TrackLocalStaticRTP
	trackListChan chan<- []*webrtc.TrackLocalStaticRTP
}

type hubRequestKind int

const (
	addTrack hubRequestKind = iota + 1
	removeTrack
	getTrackList
)
