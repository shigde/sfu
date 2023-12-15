package rtp

import (
	"context"
	"errors"
	"fmt"

	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/metric"
	"github.com/shigde/sfu/internal/rtp/stats"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/slog"
)

var ErrIceGatheringInterruption = errors.New("getting ice gathering interrupted")

type Endpoint struct {
	peerConnection peerConnection
	receiver       *receiver
	gatherComplete <-chan struct{}
	initComplete   chan struct{}
	closed         chan struct{}
	onEstablished  func()
	statsRegistry  *stats.Registry
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

func (c *Endpoint) AddTrack(track webrtc.TrackLocal) {
	slog.Debug("endpoint: Add Track", "streamId", track.StreamID(), "trackId", track.ID(), "purpose", track.Kind())
	if has := c.hasTrack(track); !has {
		slog.Debug("rtp.connection: Add Track to connection", "streamId", track.StreamID(), "trackId", track.ID(), "purpose", track.Kind(), "signalState", c.peerConnection.SignalingState().String())
		var sender *webrtc.RTPSender
		var err error
		if sender, err = c.peerConnection.AddTrack(track); err == nil {
			slog.Error("rtp.connection: Add Track to connection", "err", err, "streamId", track.StreamID(), "trackId", track.ID(), "purpose", track.Kind(), "signalState", c.peerConnection.SignalingState().String())
			return
		}
		// collect stats
		if c.statsRegistry != nil {
			labels := metric.Labels{
				metric.Stream:    track.StreamID(),
				metric.TrackId:   track.ID(),
				metric.TrackKind: track.Kind().String(),
				metric.TrackType: "guest",
				metric.Direction: "egress",
			}
			for _, param := range sender.GetParameters().Encodings {
				if err = c.statsRegistry.StartWorker(labels, param.SSRC); err != nil {
					slog.Error("endpoint: start stats worker", "err", err, "ssrc", param.SSRC)
				}
			}
		}
	}
}

func (c *Endpoint) RemoveTrack(track webrtc.TrackLocal) {
	slog.Debug("rtp.endpoint: Remove Track", "streamId", track.StreamID(), "trackId", track.ID(), "purpose", track.Kind())
	if sender, has := c.getSender(track); has {
		slog.Debug("rtp.endpoint: Remove Track from connection", "streamId", track.StreamID(), "trackId", track.ID(), "purpose", track.Kind())
		if err := c.peerConnection.RemoveTrack(sender); err != nil {
			slog.Error("rtp.endpoint: Remove Track from connection", "err", err, "streamId", track.StreamID(), "trackId", track.ID(), "purpose", track.Kind())
		}
		if c.statsRegistry != nil {
			for _, param := range sender.GetParameters().Encodings {
				c.statsRegistry.StopWorker(param.SSRC)
			}
		}
	}
}

func (c *Endpoint) Close() error {
	if err := c.peerConnection.Close(); err != nil {
		return fmt.Errorf("closing peer connection: %w", err)
	}
	if c.statsRegistry != nil {
		c.statsRegistry.StopAllWorker()
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
	Close() error
}
