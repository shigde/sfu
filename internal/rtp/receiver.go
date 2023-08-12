package rtp

import (
	"context"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/telemetry"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/slog"
)

type receiver struct {
	sync.RWMutex
	id         uuid.UUID
	streams    map[string]*localStream
	dispatcher TrackDispatcher
	quit       chan struct{}
}

func newReceiver(d TrackDispatcher) *receiver {
	streams := make(map[string]*localStream)
	quit := make(chan struct{})
	return &receiver{sync.RWMutex{}, uuid.New(), streams, d, quit}
}

func (r *receiver) onTrack(remoteTrack *webrtc.TrackRemote, rtpReceiver *webrtc.RTPReceiver) {
	slog.Debug("rtp.receiver: on track")
	ctx, span := otel.Tracer(tracerName).Start(context.Background(), "rtp.receiver: on track")
	defer span.End()

	stream := r.getStream(remoteTrack.StreamID())

	if strings.HasPrefix(remoteTrack.Codec().RTPCodecCapability.MimeType, "audio") {
		slog.Debug("rtp.receiver: on audio track")
		if err := stream.writeAudioRtp(ctx, remoteTrack, rtpReceiver); err != nil {
			slog.Error("rtp.receiver: on audio track", "err", err)
			telemetry.RecordError(span, err)
			// stop handler goroutine because error
			return
		}
		r.dispatcher.DispatchAddTrack(stream.audioTrack)
	}

	if strings.HasPrefix(remoteTrack.Codec().RTPCodecCapability.MimeType, "video") {
		slog.Debug("rtp.receiver: on video track")
		if err := stream.writeVideoRtp(ctx, remoteTrack, rtpReceiver); err != nil {
			slog.Error("rtp.receiver: on video track", "err", err)
			telemetry.RecordError(span, err)
			// stop handler goroutine because error
			return
		}
		r.dispatcher.DispatchAddTrack(stream.videoTrack)
	}
}

func (r *receiver) getStream(id string) *localStream {
	r.Lock()
	defer r.Unlock()
	stream, ok := r.streams[id]
	if !ok {
		stream = newLocalStream(id, r.dispatcher, r.quit)
		r.streams[id] = stream
	}
	return stream
}

func (r *receiver) stop() {
	slog.Info("receiver: stop", "id", r.id)
	select {
	case <-r.quit:
		slog.Warn("receiver: the receiver was already closed", "id", r.id)
	default:
		close(r.quit)
		slog.Info("receiver: stopped was triggered", "id", r.id)
	}
}
