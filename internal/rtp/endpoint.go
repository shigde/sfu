package rtp

import (
	"context"
	"errors"
	"fmt"

	"github.com/pion/webrtc/v3"
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
	slog.Debug("rtp.connection: Add Track", "streamId", track.StreamID(), "trackId", track.ID(), "kind", track.Kind())
	if has := c.hasTrack(track); !has {
		slog.Debug("rtp.connection: Add Track to connection", "streamId", track.StreamID(), "trackId", track.ID(), "kind", track.Kind(), "signalState", c.peerConnection.SignalingState().String())
		if _, err := c.peerConnection.AddTrack(track); err != nil {
			slog.Error("rtp.connection: Add Track to connection", "err", err, "streamId", track.StreamID(), "trackId", track.ID(), "kind", track.Kind(), "signalState", c.peerConnection.SignalingState().String())
		}
	}
}

func (c *Endpoint) RemoveTrack(track webrtc.TrackLocal) {
	slog.Debug("rtp.connection: Remove Track", "streamId", track.StreamID(), "trackId", track.ID(), "kind", track.Kind())
	if sender, has := c.getSender(track); has {
		slog.Debug("rtp.connection: Remove Track from connection", "streamId", track.StreamID(), "trackId", track.ID(), "kind", track.Kind())
		if err := c.peerConnection.RemoveTrack(sender); err != nil {
			slog.Error("rtp.connection: Remove Track from connection", "err", err, "streamId", track.StreamID(), "trackId", track.ID(), "kind", track.Kind())
		}
	}
}

func (c *Endpoint) Close() error {
	if err := c.peerConnection.Close(); err != nil {
		return fmt.Errorf("closing peer connection: %w", err)
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
