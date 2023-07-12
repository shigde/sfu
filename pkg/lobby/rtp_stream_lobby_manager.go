package lobby

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

var errLobbyRequestTimeout = errors.New("lobby request timeout error")

type RtpStreamLobbyManager struct {
	lobbies *RtpStreamLobbyRepository
}

func NewLobbyManager() *RtpStreamLobbyManager {
	lobbies := newRtpStreamLobbyRepository()
	return &RtpStreamLobbyManager{lobbies}
}

func (m *RtpStreamLobbyManager) AccessLobby(ctx context.Context, liveStreamId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription) (struct {
	Answer       *webrtc.SessionDescription
	Resource     uuid.UUID
	RtpSessionId uuid.UUID
}, error) {
	errChan := make(chan error)
	resChan := make(chan *joinResponse)
	defer func() {
		close(errChan)
		close(resChan)
	}()

	lobby := m.lobbies.getOrCreateLobby(liveStreamId)
	joinRequest := &joinRequest{
		offer:    offer,
		user:     user,
		error:    errChan,
		response: resChan,
	}
	lobby.request <- joinRequest

	var data struct {
		Answer       *webrtc.SessionDescription
		Resource     uuid.UUID
		RtpSessionId uuid.UUID
	}

	select {
	case err := <-errChan:
		return data, fmt.Errorf("joining lobby %w", err)
	case rtpResourceData := <-resChan:
		data.Answer = rtpResourceData.answer
		data.Resource = rtpResourceData.resource
		data.RtpSessionId = rtpResourceData.RtpSessionId
		return data, nil
	case <-ctx.Done():
		return data, errLobbyRequestTimeout
	}
}
