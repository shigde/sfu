package lobby

import (
	"github.com/pion/webrtc/v3"
)

type RtpStreamLobbyManager struct {
	lobbies *RtpStreamLobbyRepository
}

func NewLobbyManager() *RtpStreamLobbyManager {
	lobbies := newRtpStreamLobbyRepository()
	return &RtpStreamLobbyManager{lobbies}
}

type Answer struct {
	sdp     *webrtc.SessionDescription
	session string
}

func (m *RtpStreamLobbyManager) AccessLobby(id string) (*RtpStreamLobby, error) {
	// The error is need for distributed lobbies later
	return m.lobbies.getOrCreateLobby(id), nil
}
