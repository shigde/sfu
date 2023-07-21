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

type LobbyManager struct {
	lobbies *RtpStreamLobbyRepository
}

type rtpEngine interface {
	NewConnection(offer webrtc.SessionDescription, _ string) (*rtp.Connection, error)
}

func NewLobbyManager(e rtpEngine) *LobbyManager {
	lobbies := newRtpStreamLobbyRepository(e)
	return &LobbyManager{lobbies}
}

func (m *LobbyManager) AccessLobby(ctx context.Context, liveStreamId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription) (struct {
	Answer       *webrtc.SessionDescription
	Resource     uuid.UUID
	RtpSessionId uuid.UUID
}, error) {
	lobby := m.lobbies.getOrCreateLobby(liveStreamId)
	joinRequest := newJoinRequest(ctx, user, offer)

	go lobby.runJoin(joinRequest)

	var data struct {
		Answer       *webrtc.SessionDescription
		Resource     uuid.UUID
		RtpSessionId uuid.UUID
	}

	select {
	case err := <-joinRequest.err:
		return data, fmt.Errorf("joining lobby: %w", err)
	case rtpResourceData := <-joinRequest.response:
		data.Answer = rtpResourceData.answer
		data.Resource = rtpResourceData.resource
		data.RtpSessionId = rtpResourceData.RtpSessionId
		return data, nil
	}
}

func (m *LobbyManager) ListenLobby(ctx context.Context, liveStreamId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription) (struct {
	Answer       *webrtc.SessionDescription
	Active       bool
	RtpSessionId uuid.UUID
}, error) {

	var data struct {
		Answer       *webrtc.SessionDescription
		Active       bool
		RtpSessionId uuid.UUID
	}

	if lobby, hasLobby := m.lobbies.getLobby(liveStreamId); hasLobby {
		joinRequest := newJoinRequest(ctx, user, offer)

		go lobby.runJoin(joinRequest)

		var data struct {
			Answer       *webrtc.SessionDescription
			Active       bool
			RtpSessionId uuid.UUID
		}

		select {
		case err := <-joinRequest.err:
			return data, fmt.Errorf("joining lobby: %w", err)
		case rtpResourceData := <-joinRequest.response:
			data.Answer = rtpResourceData.answer
			data.Active = true
			data.RtpSessionId = rtpResourceData.RtpSessionId
			return data, nil
		}
	}
	data.Active = false
	return data, nil
}
