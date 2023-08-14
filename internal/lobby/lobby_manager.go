package lobby

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/rtp"
)

const tracerName = "github.com/shigde/sfu/internal/lobby"

var errLobbyRequestTimeout = errors.New("lobby request timeout error")

type LobbyManager struct {
	lobbies *lobbyRepository
}

type rtpEngine interface {
	NewReceiverEndpoint(ctx context.Context, offer webrtc.SessionDescription, d rtp.TrackDispatcher) (*rtp.Endpoint, error)
	NewSenderEndpoint(ctx context.Context, localTracks []*webrtc.TrackLocalStaticRTP) (*rtp.Endpoint, error)
}

func NewLobbyManager(e rtpEngine) *LobbyManager {
	lobbies := newLobbyRepository(e)
	return &LobbyManager{lobbies}
}

func (m *LobbyManager) AccessLobby(ctx context.Context, liveStreamId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription) (struct {
	Answer       *webrtc.SessionDescription
	Resource     uuid.UUID
	RtpSessionId uuid.UUID
}, error) {
	lobby := m.lobbies.getOrCreateLobby(liveStreamId)
	request := newLobbyRequest(ctx, user)
	joinData := newJoinData(offer)
	request.data = joinData

	go lobby.runRequest(request)

	var answerData struct {
		Answer       *webrtc.SessionDescription
		Resource     uuid.UUID
		RtpSessionId uuid.UUID
	}

	select {
	case err := <-request.err:
		return answerData, fmt.Errorf("requesting joining lobby: %w", err)
	case rtpResourceData := <-joinData.response:
		answerData.Answer = rtpResourceData.answer
		answerData.Resource = rtpResourceData.resource
		answerData.RtpSessionId = rtpResourceData.RtpSessionId
		return answerData, nil
	}
}

func (m *LobbyManager) StartListenLobby(ctx context.Context, liveStreamId uuid.UUID, user uuid.UUID) (struct {
	Offer        *webrtc.SessionDescription
	Active       bool
	RtpSessionId uuid.UUID
}, error) {

	var answerData struct {
		Offer        *webrtc.SessionDescription
		Active       bool
		RtpSessionId uuid.UUID
	}

	if lobby, hasLobby := m.lobbies.getLobby(liveStreamId); hasLobby {
		request := newLobbyRequest(ctx, user)
		listenData := newStartListenData()
		request.data = listenData

		go lobby.runRequest(request)

		var data struct {
			Offer        *webrtc.SessionDescription
			Active       bool
			RtpSessionId uuid.UUID
		}

		select {
		case err := <-request.err:
			return data, fmt.Errorf("requesting listening lobby: %w", err)
		case rtpResourceData := <-listenData.response:
			data.Offer = rtpResourceData.offer
			data.Active = true
			data.RtpSessionId = rtpResourceData.RtpSessionId
			return data, nil
		}
	}
	answerData.Active = false
	return answerData, nil
}

func (m *LobbyManager) ListenLobby(ctx context.Context, liveStreamId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription) (struct {
	Answer       *webrtc.SessionDescription
	Active       bool
	RtpSessionId uuid.UUID
}, error) {

	var answerData struct {
		Answer       *webrtc.SessionDescription
		Active       bool
		RtpSessionId uuid.UUID
	}

	if lobby, hasLobby := m.lobbies.getLobby(liveStreamId); hasLobby {
		request := newLobbyRequest(ctx, user)
		listenData := newListenData(offer)
		request.data = listenData

		go lobby.runRequest(request)

		var data struct {
			Answer       *webrtc.SessionDescription
			Active       bool
			RtpSessionId uuid.UUID
		}

		select {
		case err := <-request.err:
			return data, fmt.Errorf("requesting listening lobby: %w", err)
		case rtpResourceData := <-listenData.response:
			data.Active = true
			data.RtpSessionId = rtpResourceData.RtpSessionId
			return data, nil
		}
	}
	answerData.Active = false
	return answerData, nil
}

func (m *LobbyManager) LeaveLobby(ctx context.Context, liveStreamId uuid.UUID, userId uuid.UUID) (bool, error) {
	return true, nil
}
