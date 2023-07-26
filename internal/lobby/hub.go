package lobby

import (
	"errors"

	"github.com/pion/webrtc/v3"
	"golang.org/x/exp/slog"
)

var errHubAlreadyClosed = errors.New("hub was already closed")

type hub struct {
	sessionRepo *sessionRepository
	onTrack     chan *webrtc.TrackLocalStaticRTP
	quit        chan struct{}
}

func newHub(sessionRepo *sessionRepository) *hub {
	quit := make(chan struct{})
	onTrack := make(chan *webrtc.TrackLocalStaticRTP)
	hub := &hub{
		sessionRepo,
		onTrack,
		quit,
	}
	hub.run()
	return hub
}

func (h *hub) run() {
	slog.Info("lobby.hub: run")
	for {
		select {
		case track := <-h.onTrack:
			h.addTrackToSessions(track)
		case <-h.quit:
			slog.Info("lobby.hub: close hub")
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
		close(s.quit)
		slog.Info("lobby.hub: stopped was triggered")
	}
	return nil
}

func (h *hub) addTrackToSessions(track *webrtc.TrackLocalStaticRTP) {
	h.sessionRepo.Iter(func(s *session) {
		// der kann theoretisch geschlossen sein
		s.onRemoteTrack <- track
	})
}
