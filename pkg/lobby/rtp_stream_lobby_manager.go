package lobby

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

type RtpStreamLobbyManager struct {
	lobbies *RtpStreamLobbyRepository
}

func NewLobbyManager() *RtpStreamLobbyManager {
	lobbies := newRtpStreamLobbyRepository()
	return &RtpStreamLobbyManager{lobbies}
}

func (m *RtpStreamLobbyManager) AccessLobby(liveStreamId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription) (struct {
	Answer       *webrtc.SessionDescription
	Resource     uuid.UUID
	RtpSessionId uuid.UUID
}, error) {
	// The error is need for distributed lobbies later
	errChan := make(chan error)
	resChan := make(chan *RtpResourceData)
	defer func() {
		close(errChan)
		close(resChan)
	}()

	lobby := m.lobbies.getOrCreateLobby(liveStreamId)

	// das ist quatsch
	go func(user uuid.UUID, offer *webrtc.SessionDescription) {
		resource, err := lobby.Join(user, offer)
		if err != nil {
			errChan <- fmt.Errorf("joining lobby %w", err)
			return
		}
		resChan <- resource
	}(user, offer)

	var data struct {
		Answer       *webrtc.SessionDescription
		Resource     uuid.UUID
		RtpSessionId uuid.UUID
	}

	select {
	case err := <-errChan:
		return data, err
	case rtpResourceData := <-resChan:
		data.Answer = rtpResourceData.Answer
		data.Resource = rtpResourceData.Resource
		data.RtpSessionId = rtpResourceData.RtpSessionId
		return data, nil
	}

}
