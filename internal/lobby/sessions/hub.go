package sessions

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/metric"
	"github.com/shigde/sfu/internal/rtp"
	"golang.org/x/exp/slog"
)

var (
	errHubAlreadyClosed   = errors.New("Hub was already closed")
	errHubDispatchTimeOut = errors.New("Hub dispatch timeout")
	hubDispatchTimeout    = 3 * time.Second
)

type liveStreamSender interface {
	AddTrack(track webrtc.TrackLocal)
	RemoveTrack(track webrtc.TrackLocal)
}

type Hub struct {
	ctx           context.Context
	LiveStreamId  uuid.UUID
	sessionRepo   *SessionRepository
	sender        liveStreamSender
	reqChan       chan *hubRequest
	tracks        map[string]*rtp.TrackInfo   // trackID --> TrackInfo
	metricNodes   map[string]metric.GraphNode // sessionId --> metric Node
	hubMetricNode metric.GraphNode
}

func NewHub(ctx context.Context, sessionRepo *SessionRepository, liveStream uuid.UUID, sender liveStreamSender) *Hub {
	tracks := make(map[string]*rtp.TrackInfo)
	metricNodes := make(map[string]metric.GraphNode)
	requests := make(chan *hubRequest)
	hubMetricNode := metric.GraphNodeUpdate(metric.BuildNode(liveStream.String(), liveStream.String(), "Hub"))
	hub := &Hub{
		ctx,
		liveStream,
		sessionRepo,
		sender,
		requests,
		tracks,
		metricNodes,
		hubMetricNode,
	}
	go hub.run()

	go func(done <-chan struct{}, node metric.GraphNode) {
		<-done
		metric.GraphNodeDelete(node)
	}(ctx.Done(), hubMetricNode)

	return hub
}

func (h *Hub) run() {
	slog.Info("lobby.Hub: run")
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
			case muteTrack:
				h.onMuteTrack(trackEvent)
			}
		case <-h.ctx.Done():
			slog.Info("lobby.Hub: closed Hub")
			return
		}
	}
}

func (h *Hub) DispatchAddTrack(ctx context.Context, track *rtp.TrackInfo) {
	select {
	case h.reqChan <- &hubRequest{ctx: ctx, kind: addTrack, track: track}:
		slog.Debug("lobby.Hub: dispatch add track", "streamId", track.GetTrackLocal().StreamID(), "track", track.GetTrackLocal().ID(), "kind", track.GetTrackLocal().Kind(), "purpose", track.Purpose.ToString())
	case <-h.ctx.Done():
		slog.Warn("lobby.Hub: dispatch add track even on closed Hub")
	case <-time.After(hubDispatchTimeout):
		slog.Error("lobby.Hub: dispatch add track - interrupted because dispatch timeout")
	}
}

func (h *Hub) DispatchRemoveTrack(ctx context.Context, track *rtp.TrackInfo) {
	select {
	case h.reqChan <- &hubRequest{ctx: ctx, kind: removeTrack, track: track}:
		slog.Debug("lobby.Hub: dispatch remove track", "streamId", track.GetTrackLocal().StreamID(), "track", track.GetTrackLocal().ID(), "kind", track.GetTrackLocal().Kind(), "purpose", track.Purpose.ToString())
	case <-h.ctx.Done():
		slog.Warn("lobby.Hub: dispatch remove track even on closed Hub")
	case <-time.After(hubDispatchTimeout):
		slog.Error("lobby.Hub: dispatch remove track - interrupted because dispatch timeout")
	}
}

func (h *Hub) DispatchMuteTrack(ctx context.Context, track *rtp.TrackInfo) {
	select {
	case h.reqChan <- &hubRequest{ctx: ctx, kind: muteTrack, track: track}:
		slog.Debug("lobby.Hub: dispatch mute track", "id", track.GetId(), "purpose", track.Purpose.ToString())
	case <-h.ctx.Done():
		slog.Warn("lobby.Hub: dispatch mute track even on closed Hub")
	case <-time.After(hubDispatchTimeout):
		slog.Error("lobby.Hub: dispatch mute track - interrupted because dispatch timeout")
	}
}

// getTrackList Is called from the Egress endpoints when the connection is established.
// In ths wax the egress endpoints can receive the current tracks of the lobby
// The session set this methode as callback to the egress egress
func (h *Hub) getTrackList(ctx context.Context, sessionId uuid.UUID, filters ...filterHubTracks) ([]*rtp.TrackInfo, error) {
	var hubList []*rtp.TrackInfo
	trackListChan := make(chan []*rtp.TrackInfo)
	select {
	case h.reqChan <- &hubRequest{ctx: ctx, kind: getTrackList, trackListChan: trackListChan}:
	case <-h.ctx.Done():
		slog.Warn("lobby.Hub: get track list on closed Hub")
		return nil, errHubAlreadyClosed
	case <-time.After(hubDispatchTimeout):
		slog.Error("lobby.Hub: get track list - interrupted because dispatch timeout")
		return nil, errHubDispatchTimeOut
	}

	select {
	case hubList = <-trackListChan:
	case <-h.ctx.Done():
		slog.Warn("lobby.Hub: get track list on closed Hub")
	case <-time.After(hubDispatchTimeout):
		slog.Error("lobby.Hub: get track list - interrupted because dispatch timeout")
	}
	list := make([]*rtp.TrackInfo, 0)
	for _, track := range hubList {
		// If we have no filter, we can add the track to the list
		if len(filters) == 0 {
			list = append(list, track)
			h.increaseNodeGraphStats(sessionId.String(), rtp.EgressEndpoint, track.Purpose)
			continue
		}

		// Check filter
		canAddTrackToList := true
		for _, f := range filters {
			// If one filter not be true we will not add the track to the list
			if !f(track) {
				canAddTrackToList = false
				break
			}
		}

		if canAddTrackToList {
			list = append(list, track)
			h.increaseNodeGraphStats(sessionId.String(), rtp.EgressEndpoint, track.Purpose)
		}
	}
	return list, nil
}

func (h *Hub) onAddTrack(event *hubRequest) {
	slog.Debug("lobby.Hub: add track", "sourceSessionId", event.track.SessionId, "streamId", event.track.GetTrackLocal().StreamID(), "track", event.track.GetTrackLocal().ID(), "kind", event.track.GetTrackLocal().Kind(), "purpose", event.track.Purpose.ToString())

	h.increaseNodeGraphStats(event.track.SessionId.String(), rtp.IngressEndpoint, event.track.Purpose)
	h.hubMetricNode = metric.GraphNodeUpdateInc(h.hubMetricNode, event.track.Purpose.ToString())
	if event.track.GetPurpose() == rtp.PurposeMain {
		slog.Debug("lobby.Hub: add live track ro sender", "streamId", event.track.GetTrackLocal().StreamID(), "track", event.track.GetTrackLocal().ID(), "kind", event.track.GetTrackLocal().Kind(), "purpose", event.track.Purpose.ToString())
		h.sender.AddTrack(event.track.GetTrackLocal())
	}

	h.tracks[event.track.GetTrackLocal().ID()] = event.track
	h.sessionRepo.Iter(func(s *Session) {
		// If a session has just been created, this call blocks for seconds.
		// This is because the ice gathering sometimes takes seconds. That's why we don't block the call
		if !s.initComplete() {
			return
		}
		if filterForSession(s.Id)(event.track) {
			slog.Debug("lobby.Hub: add egress track to session", "sessionId", s.Id, "sourceSessionId", event.track.SessionId, "streamId", event.track.GetTrackLocal().StreamID(), "track", event.track.GetTrackLocal().ID(), "kind", event.track.GetTrackLocal().Kind(), "purpose", event.track.Purpose.ToString())
			s.addTrack(event.ctx, event.track)
		}
	})
}

func (h *Hub) onRemoveTrack(event *hubRequest) {
	slog.Debug("lobby.Hub: remove track", "sourceSessionId", event.track.SessionId, "streamId", event.track.GetTrackLocal().StreamID(), "track", event.track.GetTrackLocal().ID(), "kind", event.track.GetTrackLocal().Kind(), "purpose", event.track.Purpose.ToString())

	h.hubMetricNode = metric.GraphNodeUpdateDec(h.hubMetricNode, event.track.Purpose.ToString())
	h.decreaseNodeGraphStats(event.track.SessionId.String(), rtp.IngressEndpoint, event.track.Purpose)

	if event.track.GetPurpose() == rtp.PurposeMain {
		h.sender.RemoveTrack(event.track.GetTrackLocal())
	}

	if _, ok := h.tracks[event.track.GetTrackLocal().ID()]; ok {
		delete(h.tracks, event.track.GetTrackLocal().ID())
	}

	h.sessionRepo.Iter(func(s *Session) {
		// If a session has just been created, this call blocks for seconds.
		// This is because the ice gathering sometimes takes seconds. That's why we don't block the call
		if !s.initComplete() {
			return
		}
		if filterForSession(s.Id)(event.track) {
			slog.Debug("lobby.Hub: remove egress track from session", "sessionId", s.Id, "sourceSessionId", event.track.SessionId, "streamId", event.track.GetTrackLocal().StreamID(), "track", event.track.GetTrackLocal().ID(), "kind", event.track.GetTrackLocal().Kind(), "purpose", event.track.Purpose.ToString())
			s.removeTrack(event.ctx, event.track)
			h.decreaseNodeGraphStats(s.Id.String(), rtp.EgressEndpoint, event.track.Purpose)
		}
	})
}

func (h *Hub) onGetTrackList(event *hubRequest) {
	list := make([]*rtp.TrackInfo, 0, len(h.tracks))
	for _, track := range h.tracks {
		list = append(list, track)
	}

	select {
	case event.trackListChan <- list:
	case <-h.ctx.Done():
		slog.Warn("lobby.Hub: onGetTrackList on closed Hub")
	case <-time.After(hubDispatchTimeout):
		slog.Error("lobby.Hub: onGetTrackList - interrupted because dispatch timeout")
	}
}

func (h *Hub) onMuteTrack(event *hubRequest) {
	slog.Debug("lobby.Hub: mute track", "sourceSessionId", event.track.SessionId, "streamId", "purpose", event.track.Purpose.ToString())
	h.sessionRepo.Iter(func(s *Session) {
		if filterForSession(s.Id)(event.track) {
			slog.Debug("lobby.Hub: mute egress track from session", "sessionId", s.Id, "sourceSessionId", event.track.SessionId, event.track.Purpose.ToString())
			go func(session *Session) {
				// If a session has just been created, this call blocks for seconds.
				// This is because the ice gathering sometimes takes seconds. That's why we don't block the call
				session.muteTrack(event.ctx, event.track)
			}(s)
		}
	})
}

func (h *Hub) increaseNodeGraphStats(sessionId string, endpointType rtp.EndpointType, purpose rtp.Purpose) {
	index := endpointType.ToString() + sessionId
	metricNode, ok := h.metricNodes[index]
	if !ok {
		// if metric not found create the egress and the edge from egress to lobby
		metricNode = metric.BuildNode(sessionId, h.LiveStreamId.String(), endpointType.ToString())
		metric.GraphAddEdge(sessionId, h.LiveStreamId.String(), endpointType.ToString())
	}
	h.metricNodes[index] = metric.GraphNodeUpdateInc(metricNode, purpose.ToString())
}

func (h *Hub) decreaseNodeGraphStats(sessionId string, endpointType rtp.EndpointType, purpose rtp.Purpose) {
	index := endpointType.ToString() + sessionId
	if metricNode, ok := h.metricNodes[index]; ok {
		metricNode = metric.GraphNodeUpdateDec(metricNode, purpose.ToString())
		if metricNode.Tracks == 0 && metricNode.MainTracks == 0 {
			metric.GraphDeleteEdge(sessionId, h.LiveStreamId.String(), endpointType.ToString())
			delete(h.metricNodes, index)
			return
		}
		h.metricNodes[index] = metricNode
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
