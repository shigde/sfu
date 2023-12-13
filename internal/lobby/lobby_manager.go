package lobby

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/rtp"
	"github.com/shigde/sfu/internal/storage"
	"golang.org/x/exp/slog"
)

const tracerName = "github.com/shigde/sfu/internal/lobby"

type LobbyManager struct {
	lobbies               *lobbyRepository
	lobbyGarbageCollector chan<- uuid.UUID
}

type rtpEngine interface {
	EstablishEgressEndpoint(ctx context.Context, sessionId uuid.UUID, offer webrtc.SessionDescription, d rtp.TrackDispatcher, stateHandler rtp.StateEventHandler) (*rtp.Endpoint, error)

	EstablishIngressEndpoint(ctx context.Context, sessionId uuid.UUID, localTracks []webrtc.TrackLocal, stateHandler rtp.StateEventHandler) (*rtp.Endpoint, error)

	EstablishStaticEgressEndpoint(ctx context.Context, sessionId uuid.UUID, offer webrtc.SessionDescription, options ...rtp.EndpointOption) (*rtp.Endpoint, error)
}

func NewLobbyManager(storage storage.Storage, e rtpEngine) *LobbyManager {
	lobbies := newLobbyRepository(storage, e)
	lobbyGarbageCollector := make(chan uuid.UUID)
	go func() {
		for id := range lobbyGarbageCollector {
			if ok := lobbies.delete(context.Background(), id); !ok {
				slog.Warn("lobby could not delete", "lobby", id)
			}
		}
	}()
	return &LobbyManager{lobbies, lobbyGarbageCollector}
}

func (m *LobbyManager) CreateLobbyIngressEndpoint(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription) (struct {
	Answer       *webrtc.SessionDescription
	Resource     uuid.UUID
	RtpSessionId uuid.UUID
}, error) {
	var answerData struct {
		Answer       *webrtc.SessionDescription
		Resource     uuid.UUID
		RtpSessionId uuid.UUID
	}

	lobby, err := m.lobbies.getOrCreateLobby(ctx, lobbyId, m.lobbyGarbageCollector)
	if err != nil {
		return answerData, fmt.Errorf("getting or creating lobby: %w", err)
	}

	request := newLobbyRequest(ctx, user)
	ingressData := newIngressEndpointData(offer)
	request.data = ingressData

	go lobby.runRequest(request)

	select {
	case err := <-request.err:
		return answerData, fmt.Errorf("requesting joining lobby: %w", err)
	case rtpResourceData := <-ingressData.response:
		answerData.Answer = rtpResourceData.answer
		answerData.Resource = rtpResourceData.resource
		answerData.RtpSessionId = rtpResourceData.RtpSessionId
		return answerData, nil
	}
}

func (m *LobbyManager) InitLobbyEgressEndpoint(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID) (struct {
	Offer        *webrtc.SessionDescription
	Active       bool
	RtpSessionId uuid.UUID
}, error) {

	var answerData struct {
		Offer        *webrtc.SessionDescription
		Active       bool
		RtpSessionId uuid.UUID
	}

	if lobby, hasLobby := m.lobbies.getLobby(lobbyId); hasLobby {
		request := newLobbyRequest(ctx, user)
		egressData := newInitEgressEndpointData()
		request.data = egressData

		go lobby.runRequest(request)

		var data struct {
			Offer        *webrtc.SessionDescription
			Active       bool
			RtpSessionId uuid.UUID
		}

		select {
		case err := <-request.err:
			return data, fmt.Errorf("requesting listening lobby: %w", err)
		case rtpResourceData := <-egressData.response:
			data.Offer = rtpResourceData.offer
			data.Active = true
			data.RtpSessionId = rtpResourceData.RtpSessionId
			return data, nil
		}
	}
	answerData.Active = false
	return answerData, nil
}

func (m *LobbyManager) FinalCreateLobbyEgressEndpoint(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription) (struct {
	Answer       *webrtc.SessionDescription
	Active       bool
	RtpSessionId uuid.UUID
}, error) {

	var answerData struct {
		Answer       *webrtc.SessionDescription
		Active       bool
		RtpSessionId uuid.UUID
	}

	if lobby, hasLobby := m.lobbies.getLobby(lobbyId); hasLobby {
		request := newLobbyRequest(ctx, user)
		listenData := newFinalCreateEgressEndpointData(offer)
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

func (m *LobbyManager) CreateMainStreamLobbyEgressEndpoint(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription) (struct {
	Answer       *webrtc.SessionDescription
	RtpSessionId uuid.UUID
}, error) {

	var answerData struct {
		Answer       *webrtc.SessionDescription
		RtpSessionId uuid.UUID
	}

	if lobby, hasLobby := m.lobbies.getLobby(lobbyId); hasLobby {
		request := newLobbyRequest(ctx, user)
		egressData := newMainEgressEndpointData(offer)
		request.data = egressData

		go lobby.runRequest(request)

		select {
		case err := <-request.err:
			return answerData, fmt.Errorf("requesting joining lobby: %w", err)
		case rtpResourceData := <-egressData.response:
			answerData.Answer = rtpResourceData.answer
			answerData.RtpSessionId = rtpResourceData.RtpSessionId
			return answerData, nil
		}
	}
	return answerData, errLobbyNotFound
}

func (m *LobbyManager) LeaveLobby(ctx context.Context, lobbyId uuid.UUID, userId uuid.UUID) (bool, error) {
	if lobby, hasLobby := m.lobbies.getLobby(lobbyId); hasLobby {
		request := newLobbyRequest(ctx, userId)
		leaveData := newLeaveData()
		request.data = leaveData

		go lobby.runRequest(request)

		select {
		case err := <-request.err:
			return false, err
		case success := <-leaveData.response:
			return success, nil
		}
	}
	return false, nil
}

func (m *LobbyManager) StartLiveStream(
	ctx context.Context,
	lobbyId uuid.UUID,
	key string,
	rtmpUrl string,
	userId uuid.UUID,
) error {
	if lobby, hasLobby := m.lobbies.getLobby(lobbyId); hasLobby {
		request := newLobbyRequest(ctx, userId)
		startData := newLiveStreamStart(key, rtmpUrl)
		request.data = startData
		go lobby.runRequest(request)

		select {
		case err := <-request.err:
			return err
		case res := <-startData.response:
			m.lobbies.setLobbyLive(ctx, lobbyId, res)
			return nil
		}
	}
	return nil
}

func (m *LobbyManager) StopLiveStream(
	ctx context.Context,
	lobbyId uuid.UUID,
	userId uuid.UUID,
) error {
	if lobby, hasLobby := m.lobbies.getLobby(lobbyId); hasLobby {
		request := newLobbyRequest(ctx, userId)
		stopData := newLiveStreamStop()
		request.data = stopData
		go lobby.runRequest(request)

		select {
		case err := <-request.err:
			return err
		case _ = <-stopData.response:
			m.lobbies.setLobbyLive(ctx, lobbyId, false)
			return nil
		}
	}
	return nil
}
