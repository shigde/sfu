package mocks

import "github.com/pion/webrtc/v3"

type LiveSenderMock struct {
	Tracks map[string]webrtc.TrackLocal
}

func NewLiveSender() *LiveSenderMock {
	tracks := make(map[string]webrtc.TrackLocal)
	return &LiveSenderMock{
		Tracks: tracks,
	}
}
func (sf *LiveSenderMock) AddTrack(track webrtc.TrackLocal) {
	sf.Tracks[track.ID()] = track
}

func (sf *LiveSenderMock) RemoveTrack(track webrtc.TrackLocal) {
	delete(sf.Tracks, track.ID())
}
func (sf *LiveSenderMock) GetTracks() []webrtc.TrackLocal {
	tracks := make([]webrtc.TrackLocal, len(sf.Tracks))
	for _, track := range sf.Tracks {
		tracks = append(tracks, track)
	}
	return tracks
}
