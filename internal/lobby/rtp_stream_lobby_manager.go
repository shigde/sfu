package lobby

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/rtp"
)

var errLobbyRequestTimeout = errors.New("lobby request timeout error")

type RtpStreamLobbyManager struct {
	lobbies *RtpStreamLobbyRepository
}

type rtpEngine interface {
	NewConnection(offer webrtc.SessionDescription, _ string) (*rtp.Connection, error)
}

func NewLobbyManager(e rtpEngine) *RtpStreamLobbyManager {
	lobbies := newRtpStreamLobbyRepository(e)
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
		ctx:      ctx,
	}
	lobby.request <- joinRequest

	var data struct {
		Answer       *webrtc.SessionDescription
		Resource     uuid.UUID
		RtpSessionId uuid.UUID
	}

	select {
	case err := <-errChan:
		return data, fmt.Errorf("joining lobby: %w", err)
	case rtpResourceData := <-resChan:
		data.Answer = rtpResourceData.answer
		data.Resource = rtpResourceData.resource
		data.RtpSessionId = rtpResourceData.RtpSessionId
		return data, nil
	}
}
