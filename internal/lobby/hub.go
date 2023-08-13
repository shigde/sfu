package lobby

import (
	"errors"
	"time"

	"github.com/pion/webrtc/v3"
	"golang.org/x/exp/slog"
)

var (
	errHubAlreadyClosed   = errors.New("hub was already closed")
	errHubDispatchTimeOut = errors.New("hub dispatch timeout")
	hubDispatchTimeout    = 3 * time.Second
)

type hub struct {
	sessionRepo *sessionRepository
	reqChan     chan *hubRequest
	tracks      map[string]*webrtc.TrackLocalStaticRTP
	quit        chan struct{}
}

func newHub(sessionRepo *sessionRepository) *hub {
	quit := make(chan struct{})
	tracks := make(map[string]*webrtc.TrackLocalStaticRTP)
	requests := make(chan *hubRequest)
	hub := &hub{
		sessionRepo,
		requests,
		tracks,
		quit,
	}
	go hub.run()
	return hub
}

func (h *hub) run() {
	slog.Info("lobby.hub: run")
	for {
		select {
		case trackEvent := <-h.reqChan:
			switch trackEvent.kind {
			case addTrack:
				h.onAddTrack(trackEvent)
			case removeTrack:
				h.onRemoveTrack(trackEvent)
			case getTrackList:
				h.onGetTrackList(trackEvent)
			}
		case <-h.quit:
			slog.Info("lobby.hub: closed hub")
			return
		}
	}
}

func (h *hub) DispatchAddTrack(track *webrtc.TrackLocalStaticRTP) {
	select {
	case h.reqChan <- &hubRequest{kind: addTrack, track: track}:
	case <-h.quit:
		slog.Warn("lobby.hub: dispatch add track even on closed hub")
	case <-time.After(hubDispatchTimeout):
		slog.Error("lobby.hub: dispatch add track - interrupted because dispatch timeout")
	}
}

func (h *hub) DispatchRemoveTrack(track *webrtc.TrackLocalStaticRTP) {
	select {
	case h.reqChan <- &hubRequest{kind: removeTrack, track: track}:
	case <-h.quit:
		slog.Warn("lobby.hub: dispatch remove track even on closed hub")
	case <-time.After(hubDispatchTimeout):
		slog.Error("lobby.hub: dispatch remove track - interrupted because dispatch timeout")
	}
}

func (h *hub) getTrackList() ([]*webrtc.TrackLocalStaticRTP, error) {
	var list []*webrtc.TrackLocalStaticRTP
	trackListChan := make(chan []*webrtc.TrackLocalStaticRTP)
	select {
	case h.reqChan <- &hubRequest{kind: getTrackList, trackListChan: trackListChan}:
	case <-h.quit:
		slog.Warn("lobby.hub: get track list on closed hub")
		return nil, errHubAlreadyClosed
	case <-time.After(hubDispatchTimeout):
		slog.Error("lobby.hub: get track list - interrupted because dispatch timeout")
		return nil, errHubDispatchTimeOut
	}

	select {
	case list = <-trackListChan:
	case <-h.quit:
		slog.Warn("lobby.hub: get track list on closed hub")
	case <-time.After(hubDispatchTimeout):
		slog.Error("lobby.hub: get track list - interrupted because dispatch timeout")
	}

	return list, nil
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

func (h *hub) onAddTrack(event *hubRequest) {
	h.tracks[event.track.ID()] = event.track
	h.sessionRepo.Iter(func(s *session) {
		s.addTrack(event.track)
	})
}

func (h *hub) onRemoveTrack(event *hubRequest) {
	if _, ok := h.tracks[event.track.ID()]; ok {
		delete(h.tracks, event.track.ID())
	}
	h.sessionRepo.Iter(func(s *session) {
		s.removeTrack(event.track)
	})
}

func (h *hub) onGetTrackList(event *hubRequest) {
	list := make([]*webrtc.TrackLocalStaticRTP, len(h.tracks))
	for _, track := range h.tracks {
		list = append(list, track)
	}

	select {
	case event.trackListChan <- list:
	case <-h.quit:
		slog.Warn("lobby.hub: onGetTrackList on closed hub")
	case <-time.After(hubDispatchTimeout):
		slog.Error("lobby.hub: onGetTrackList - interrupted because dispatch timeout")
	}
}
