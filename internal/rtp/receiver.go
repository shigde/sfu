package rtp

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pion/interceptor/pkg/stats"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/telemetry"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/slog"
)

type receiver struct {
	sync.RWMutex
	id         uuid.UUID
	streams    map[string]Stream
	dispatcher TrackDispatcher
	trackInfos map[string]*TrackInfo
	stats      statsGetter
	quit       chan struct{}
}

type statsGetter interface {
	getStatsGetter(id string) (stats.Getter, bool)
}

func newReceiver(sessionId uuid.UUID, d TrackDispatcher, s statsGetter, trackInfos map[string]*TrackInfo) *receiver {
	streams := make(map[string]Stream)
	quit := make(chan struct{})
	return &receiver{sync.RWMutex{}, sessionId, streams, d, trackInfos, s, quit}
}

func (r *receiver) onTrack(remoteTrack *webrtc.TrackRemote, rtpReceiver *webrtc.RTPReceiver) {
	slog.Debug("rtp.receiver: on track")
	ctx, span := otel.Tracer(tracerName).Start(context.Background(), "rtp.receiver: on track")
	defer span.End()

	stream := r.getStream(r.id, remoteTrack.StreamID(), remoteTrack.ID())

	go func(track *webrtc.TrackRemote) {
		statsGetter, ok := r.stats.getStatsGetter("r.statsId")
		if !ok {
			return
		}
		for {
			statsGetter.Get(uint32(track.SSRC()))
			//fmt.Printf("Stats for: %s\n", remoteTrack.Codec().MimeType)
			//fmt.Println(stats.InboundRTPStreamStats)

			time.Sleep(time.Second * 5)
		}
	}(remoteTrack)

	if strings.HasPrefix(remoteTrack.Codec().RTPCodecCapability.MimeType, "audio") {
		slog.Debug("rtp.receiver: on audio track")
		if err := stream.writeAudioRtp(ctx, remoteTrack, rtpReceiver); err != nil {
			slog.Error("rtp.receiver: on audio track", "err", err)
			telemetry.RecordError(span, err)
			// stop handler goroutine because error
			return
		}
		r.dispatcher.DispatchAddTrack(newTrackInfo(r.id, stream.getAudioTrack(), stream.getLiveAudioTrack(), stream.getKind()))
	}

	if strings.HasPrefix(remoteTrack.Codec().RTPCodecCapability.MimeType, "video") {
		slog.Debug("rtp.receiver: on video track")
		if err := stream.writeVideoRtp(ctx, remoteTrack, rtpReceiver); err != nil {
			slog.Error("rtp.receiver: on video track", "err", err)
			telemetry.RecordError(span, err)
			// stop handler goroutine because error
			return
		}
		r.dispatcher.DispatchAddTrack(newTrackInfo(r.id, stream.getVideoTrack(), stream.getLiveVideoTrack(), stream.getKind()))
	}
}
func (r *receiver) getTrackInfo(streamID string, trackId string) *TrackInfo {
	mid := fmt.Sprintf("%s %s", streamID, trackId)
	info, found := r.trackInfos[mid]
	if !found {
		info = &TrackInfo{SessionId: r.id, Kind: TrackInfoKindGuest}
		r.trackInfos[mid] = info
	}
	return info
}

func (r *receiver) getStream(sessionId uuid.UUID, streamId string, trackId string) Stream {
	r.Lock()
	defer r.Unlock()
	info := r.getTrackInfo(streamId, trackId)
	stream, ok := r.streams[streamId]
	if !ok {
		switch info.Kind {
		case TrackInfoKindGuest:
			stream = newLocalStream(streamId, sessionId, r.dispatcher, info.Kind, r.quit)
		case TrackInfoKindMain:
			stream = newLiveStream(streamId, sessionId, r.dispatcher, info.Kind, r.quit)
		default:
			stream = newLocalStream(streamId, sessionId, r.dispatcher, info.Kind, r.quit)
		}
		r.streams[streamId] = stream
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
