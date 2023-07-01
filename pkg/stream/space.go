package stream

import (
	"fmt"

	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/pkg/lobby"
)

type lobbyAccessor interface {
	AccessLobby(id string) (*lobby.RtpStreamLobby, error)
}

type Space struct {
	Id             string `json:"Id"`
	LiveStreamRepo *LiveStreamRepository
	lobby          lobbyAccessor
}

func newSpace(id string, lobby lobbyAccessor) *Space {
	repo := NewLiveStreamRepository()
	return &Space{Id: id, LiveStreamRepo: repo, lobby: lobby}
}

func (s *Space) EnterLobby(sdp *webrtc.SessionDescription, stream *LiveStream, userId string, role string) (*webrtc.SessionDescription, error) {
	lobbySpace, err := s.lobby.AccessLobby(s.Id)
	if err != nil {
		return nil, fmt.Errorf("creating lobby: %w", err)
	}
	offer := lobby.Offer{
		stream.Id,
		userId,
		sdp,
		role,
	}

	// @TODO run this lobby as goroutine
	state, errFromLobby := lobbySpace.Join(offer)

	// @TODO: Dead log!! This will be fixed in next commit
	select {
	case err, _ := <-errFromLobby:
		return nil, fmt.Errorf("reading from ReadWriter a: %w", err)
	case answer, ok := <-state:
		if ok {
			return answer.Sdp, nil
		}
		// channel closed Lobby closed!
		return nil, fmt.Errorf("reading from ReadWriter a: %w", err)
	}

}
