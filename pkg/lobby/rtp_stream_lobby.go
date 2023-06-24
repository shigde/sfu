package lobby

import "github.com/pion/webrtc/v3"

type RtpStreamLobby struct {
	Id   string
	repo *rtpStreamRepository
}

func newRtpStreamLobby(id string) *RtpStreamLobby {
	repo := newRtpStreamRepository()
	return &RtpStreamLobby{Id: id, repo: repo}
}

type Offer struct {
	StreamId string
	UserId   string
	Sdp      *webrtc.SessionDescription
	Role     string
}

type LobbyState struct {
	Sdp *webrtc.SessionDescription
}

func (l *RtpStreamLobby) Join(_ Offer) (<-chan *LobbyState, <-chan error) {
	return nil, nil
}
