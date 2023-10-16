package lobby

type hubRequest struct {
	kind          hubRequestKind
	track         HubTrack
	trackListChan chan<- []HubTrack
}

type hubRequestKind int

const (
	addTrack hubRequestKind = iota + 1
	removeTrack
	getTrackList
)
