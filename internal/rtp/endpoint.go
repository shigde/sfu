package rtp

import (
	"context"
	"errors"
	"fmt"

	"github.com/pion/dtls/v2"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/metric"
	"github.com/shigde/sfu/internal/rtp/stats"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/slog"
)

var ErrIceGatheringInterruption = errors.New("getting ice gathering interrupted")

type Endpoint struct {
	sessionId      string
	liveStreamId   string
	endpointType   EndpointType
	peerConnection peerConnection
	receiver       *receiver
	gatherComplete <-chan struct{}
	initComplete   chan struct{}
	closed         chan struct{}
	statsRegistry  *stats.Registry
	// Endpoint Optional
	onChannel                  func(dc *webrtc.DataChannel)
	onEstablished              func()
	onNegotiationNeeded        func(offer webrtc.SessionDescription)
	onICEConnectionStateChange func(webrtc.ICEConnectionState)
	initTracks                 []*initTrack
	dispatcher                 TrackDispatcher
}

func (c *Endpoint) GetLocalDescription(ctx context.Context) (*webrtc.SessionDescription, error) {
	// block until ice gathering is complete before return local sdp
	// all ice candidates should be part of the answer
	_, span := otel.Tracer(tracerName).Start(ctx, "endpoint:GetLocalDescription")
	defer span.End()
	select {
	case <-c.gatherComplete:
		return c.peerConnection.LocalDescription(), nil
	case <-ctx.Done():
		return nil, ErrIceGatheringInterruption
	}
}
func (c *Endpoint) SetAnswer(sdp *webrtc.SessionDescription) error {
	return c.peerConnection.SetRemoteDescription(*sdp)
}

func (c *Endpoint) SetInitComplete() {
	select {
	case <-c.initComplete:
	default:
		close(c.initComplete)
		<-c.initComplete
	}
}

func (c *Endpoint) hasTrack(track webrtc.TrackLocal) bool {
	slog.Debug("rtp.connection: has Tracks")
	rtpSenderList := c.peerConnection.GetSenders()
	for _, rtpSender := range rtpSenderList {
		if rtpTrack := rtpSender.Track(); rtpTrack != nil {
			if rtpTrack.ID() == track.ID() {
				return true
			}
		}
	}
	return false
}

func (c *Endpoint) getSender(track webrtc.TrackLocal) (*webrtc.RTPSender, bool) {
	slog.Debug("rtp.connection: has Tracks")
	rtpSenderList := c.peerConnection.GetSenders()
	for _, rtpSender := range rtpSenderList {
		if rtpTrack := rtpSender.Track(); rtpTrack != nil {
			if rtpTrack.ID() == track.ID() {
				return rtpSender, true
			}
		}
	}
	return nil, false
}

func (c *Endpoint) AddTrack(track webrtc.TrackLocal, purpose Purpose) {
	slog.Debug("rtp.endpoint: add track", "streamId", track.StreamID(), "trackId", track.ID(), "kind", track.Kind(), "purpose", purpose)
	if has := c.hasTrack(track); !has {
		slog.Debug("rtp.endpoint: add track to connection", "streamId", track.StreamID(), "trackId", track.ID(), "kind", track.Kind(), "purpose", purpose, "signalState", c.peerConnection.SignalingState().String())
		var sender *webrtc.RTPSender
		var err error
		if sender, err = c.peerConnection.AddTrack(track); err != nil {
			slog.Error("rtp.endpoint: add track to connection", "err", err, "streamId", track.StreamID(), "trackId", track.ID(), "kind", track.Kind(), "purpose", purpose, "signalState", c.peerConnection.SignalingState().String())
			return
		}
		// collect stats
		if c.statsRegistry != nil {
			labels := metric.Labels{
				metric.Stream:       c.liveStreamId,
				metric.MediaStream:  track.StreamID(),
				metric.TrackId:      track.ID(),
				metric.TrackKind:    track.Kind().String(),
				metric.TrackPurpose: purpose.ToString(),
				metric.Direction:    c.endpointType.ToString(),
			}
			for _, param := range sender.GetParameters().Encodings {
				if err = c.statsRegistry.StartWorker(labels, param.SSRC); err != nil {
					slog.Error("rtp.endpoint: start stats worker", "err", err, "ssrc", param.SSRC)
				}
			}
		}
	}
}

func (c *Endpoint) RemoveTrack(track webrtc.TrackLocal) {
	slog.Debug("rtp.endpoint: remove track", "streamId", track.StreamID(), "trackId", track.ID(), "purpose", track.Kind())
	if sender, has := c.getSender(track); has {
		slog.Debug("rtp.endpoint: remove track from connection", "streamId", track.StreamID(), "trackId", track.ID(), "purpose", track.Kind())
		if err := c.peerConnection.RemoveTrack(sender); err != nil {
			slog.Error("rtp.endpoint: remove track from connection", "err", err, "streamId", track.StreamID(), "trackId", track.ID(), "purpose", track.Kind())
		}
		if c.statsRegistry != nil {
			for _, param := range sender.GetParameters().Encodings {
				c.statsRegistry.StopWorker(param.SSRC)
			}
		}
	}
}

func (c *Endpoint) Close() error {
	if c.statsRegistry != nil {
		c.statsRegistry.StopAllWorker()
	}

	if err := c.peerConnection.Close(); err != nil && !errors.Is(err, dtls.ErrConnClosed) {
		return fmt.Errorf("closing peer connection: %w", err)
	}

	if c.sessionId != "" && c.liveStreamId != "" {
		metric.GraphNodeDelete(metric.BuildNode(c.sessionId, c.liveStreamId, c.endpointType.ToString()))
	}

	return nil
}

type peerConnection interface {
	LocalDescription() *webrtc.SessionDescription
	SetRemoteDescription(desc webrtc.SessionDescription) error
	GetSenders() (result []*webrtc.RTPSender)
	AddTrack(track webrtc.TrackLocal) (*webrtc.RTPSender, error)
	RemoveTrack(sender *webrtc.RTPSender) error
	SignalingState() webrtc.SignalingState

	OnICEConnectionStateChange(f func(webrtc.ICEConnectionState))

	OnNegotiationNeeded(f func())
	Close() error
}

type initTrack struct {
	purpose Purpose
	track   webrtc.TrackLocal
}

type EndpointType int

const (
	IngressEndpoint EndpointType = iota + 1
	EgressEndpoint
)

func (et EndpointType) ToString() string {
	switch et {
	case IngressEndpoint:
		return "ingress"
	case EgressEndpoint:
		return "egress"
	default:
		return "unknown"
	}
}
