package lobby

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/rtp"
	"golang.org/x/exp/slog"
)

var (
	errHubAlreadyClosed   = errors.New("hub was already closed")
	errHubDispatchTimeOut = errors.New("hub dispatch timeout")
	hubDispatchTimeout    = 3 * time.Second
)

type hub struct {
	LiveStreamId uuid.UUID
	sessionRepo  *sessionRepository
	streamer     mainStreamer
	reqChan      chan *hubRequest
	tracks       map[string]*rtp.TrackInfo
	quit         chan struct{}
}

func newHub(sessionRepo *sessionRepository, liveStream uuid.UUID, forwarder mainStreamer, quit chan struct{}) *hub {
	tracks := make(map[string]*rtp.TrackInfo)
	requests := make(chan *hubRequest)
	hub := &hub{
		liveStream,
		sessionRepo,
		forwarder,
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

func (h *hub) DispatchAddTrack(track *rtp.TrackInfo) {
	select {
	case h.reqChan <- &hubRequest{kind: addTrack, track: track}:
		slog.Debug("lobby.hub: dispatch add track", "streamId", track.GetTrackLocal().StreamID(), "track", track.GetTrackLocal().ID(), "kind", track.GetTrackLocal().Kind(), "purpose", track.Purpose.ToString())
	case <-h.quit:
		slog.Warn("lobby.hub: dispatch add track even on closed hub")
	case <-time.After(hubDispatchTimeout):
		slog.Error("lobby.hub: dispatch add track - interrupted because dispatch timeout")
	}
}

func (h *hub) DispatchRemoveTrack(track *rtp.TrackInfo) {
	select {
	case h.reqChan <- &hubRequest{kind: removeTrack, track: track}:
		slog.Debug("lobby.hub: dispatch remove track", "streamId", track.GetTrackLocal().StreamID(), "track", track.GetTrackLocal().ID(), "kind", track.GetTrackLocal().Kind(), "purpose", track.Purpose.ToString())
	case <-h.quit:
		slog.Warn("lobby.hub: dispatch remove track even on closed hub")
	case <-time.After(hubDispatchTimeout):
		slog.Error("lobby.hub: dispatch remove track - interrupted because dispatch timeout")
	}
}

func (h *hub) getTrackList(filters ...filterHubTracks) ([]*rtp.TrackInfo, error) {
	var hubList []*rtp.TrackInfo
	trackListChan := make(chan []*rtp.TrackInfo)
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
	case hubList = <-trackListChan:
	case <-h.quit:
		slog.Warn("lobby.hub: get track list on closed hub")
	case <-time.After(hubDispatchTimeout):
		slog.Error("lobby.hub: get track list - interrupted because dispatch timeout")
	}
	list := make([]*rtp.TrackInfo, 0)
	for _, track := range hubList {
		if len(filters) == 0 {
			list = append(list, track)
		}

		for _, f := range filters {
			if f(track) {
				list = append(list, track)
			}
		}
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
	if event.track.GetPurpose() == rtp.PurposeMain {
		slog.Debug("lobby.hub: add live track ro streamer", "streamId", event.track.GetTrackLocal().StreamID(), "track", event.track.GetTrackLocal().ID(), "kind", event.track.GetTrackLocal().Kind(), "purpose", event.track.Purpose.ToString())
		h.streamer.AddTrack(event.track.GetLiveTrack())
	}

	h.tracks[event.track.GetTrackLocal().ID()] = event.track
	h.sessionRepo.Iter(func(s *session) {
		if filterForSession(s.Id)(event.track) {
			slog.Debug("lobby.hub: add egress track to session", "session", s.Id, "trackSession", event.track.SessionId, "streamId", event.track.GetTrackLocal().StreamID(), "track", event.track.GetTrackLocal().ID(), "kind", event.track.GetTrackLocal().Kind(), "purpose", event.track.Purpose.ToString())
			s.addTrack(event.track)
		}
	})
}

func (h *hub) onRemoveTrack(event *hubRequest) {
	if event.track.GetPurpose() == rtp.PurposeMain {
		h.streamer.RemoveTrack(event.track.GetLiveTrack())
		return
	}

	if _, ok := h.tracks[event.track.GetTrackLocal().ID()]; ok {
		delete(h.tracks, event.track.GetTrackLocal().ID())
	}
	h.sessionRepo.Iter(func(s *session) {
		if filterForSession(s.Id)(event.track) {
			s.removeTrack(event.track)
		}
	})
}

func (h *hub) onGetTrackList(event *hubRequest) {
	list := make([]*rtp.TrackInfo, 0, len(h.tracks))
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

type filterHubTracks func(*rtp.TrackInfo) bool

func filterForSession(sessionId uuid.UUID) filterHubTracks {
	return func(track *rtp.TrackInfo) bool {
		return sessionId.String() != track.GetSessionId().String()
	}
}
func filterForNotMain() filterHubTracks {
	return func(track *rtp.TrackInfo) bool {
		return track.Purpose != rtp.PurposeMain
	}
}
