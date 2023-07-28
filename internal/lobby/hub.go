package lobby

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"golang.org/x/exp/slog"
)

var errHubAlreadyClosed = errors.New("hub was already closed")
var hubDispachTimeout = time.Second

type hubTrackData struct {
	sessionId uuid.UUID
	streamId  string
	kind      webrtc.RTPCodecType
	track     *webrtc.TrackLocalStaticRTP
}

type hub struct {
	sessionRepo  *sessionRepository
	dispatchChan chan *hubTrackData
	quit         chan struct{}
}

func newHub(sessionRepo *sessionRepository) *hub {
	quit := make(chan struct{})
	dispatchChan := make(chan *hubTrackData)
	hub := &hub{
		sessionRepo,
		dispatchChan,
		quit,
	}
	go hub.run()
	return hub
}

func (h *hub) run() {
	slog.Info("lobby.hub: run")
	for {
		select {
		case trackData := <-h.dispatchChan:
			h.dispatchToSessions(trackData)
		case <-h.quit:
			slog.Info("lobby.hub: closed hub")
			return
		}
	}
}

func (h *hub) stop() error {
	slog.Info("lobby.hub: stop")
	select {
	case <-h.quit:
		slog.Error("lobby.sessions: the hub was already closed")
		return errHubAlreadyClosed
	default:
		close(h.quit)
		slog.Info("lobby.hub: stopped was triggered")
		<-h.quit
	}
	return nil
}

func (h *hub) getAllTracksFromSessions() []*webrtc.TrackLocalStaticRTP {
	slog.Debug("lobby.hub: getAllTracksFromSessions")
	var tracks []*webrtc.TrackLocalStaticRTP
	h.sessionRepo.Iter(func(s *session) {
		if sessionTracks := s.getTracks(); sessionTracks != nil {
			for _, s := range sessionTracks {
				tracks = append(tracks, s)
			}
		}
	})
	return tracks
}

func (h *hub) dispatchToSessions(track *hubTrackData) {
	slog.Debug("lobby.hub: dispatchToSessions")
	h.sessionRepo.Iter(func(s *session) {
		if s.Id != track.sessionId {
			select {
			case <-s.quit:
			case s.foreignTrackChan <- track:
			case <-time.After(hubDispachTimeout):
				slog.Error("lobby.hub: dispatchToSessions - interrupted because dispatch timeout", "sessionId", s.Id, "user", s.user)
			}
		}
	})
}
