package rtp

import (
	"context"
	"errors"
	"fmt"

	"github.com/pion/webrtc/v3"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/slog"
)

type Endpoint struct {
	peerConnection peerConnection
	receiver       *receiver
	sender         *sender
	AddTrackChan   <-chan *webrtc.TrackLocalStaticRTP
	gatherComplete <-chan struct{}
	closed         chan struct{}
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
		return nil, errors.New("getting answer get interrupted")
	}
}
func (c *Endpoint) SetAnswer(sdp *webrtc.SessionDescription) error {
	return c.peerConnection.SetRemoteDescription(*sdp)
}

func (c *Endpoint) hasTrack(track *webrtc.TrackLocalStaticRTP) bool {
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

func (c *Endpoint) AddTrack(track *webrtc.TrackLocalStaticRTP) {
	slog.Debug("rtp.connection: Add Track")
	if has := c.hasTrack(track); !has {
		_, err := c.peerConnection.AddTrack(track)
		slog.Debug("rtp.connection: Add Tracks to connection", "err", err)
	}
}

func (c *Endpoint) RemoveTrack(track *webrtc.TrackLocalStaticRTP) {
	slog.Debug("rtp.connection: Add Track")
	if has := c.hasTrack(track); !has {
		_, err := c.peerConnection.AddTrack(track)
		slog.Debug("rtp.connection: Add Tracks to connection", "err", err)
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
	Close() error
}