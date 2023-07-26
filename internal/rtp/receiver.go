package rtp

import (
	"strings"
	"sync"

	"github.com/pion/webrtc/v3"
	"golang.org/x/exp/slog"
)

type receiver struct {
	sync.RWMutex
	// senders []*sender
	streams  map[string]*localStream
	newTrack chan<- *webrtc.TrackLocalStaticRTP
}

func newReceiver(newTrack chan<- *webrtc.TrackLocalStaticRTP) *receiver {
	streams := make(map[string]*localStream)
	return &receiver{sync.RWMutex{}, streams, newTrack}
}

func (r *receiver) onTrack(remoteTrack *webrtc.TrackRemote, rtpReceiver *webrtc.RTPReceiver) {
	if strings.HasPrefix(remoteTrack.Codec().RTPCodecCapability.MimeType, "audio") {
		if err := r.onAudioTrack(remoteTrack, rtpReceiver); err != nil {
			slog.Error("rtp.receiver: on audio track", "err", err)
			// stop handler goroutine
			return
		}
	}

	if strings.HasPrefix(remoteTrack.Codec().RTPCodecCapability.MimeType, "video") {
		if err := r.onVideoTrack(remoteTrack, rtpReceiver); err != nil {
			slog.Error("rtp.receiver: on video track", "err", err)
			// stop handler goroutine
			return
		}
	}
}

func (r *receiver) onAudioTrack(remoteTrack *webrtc.TrackRemote, rtpReceiver *webrtc.RTPReceiver) error {
	stream := r.getStream(remoteTrack.StreamID())
	return stream.writeAudioRtp(remoteTrack, r.newTrack)
}

func (r *receiver) onVideoTrack(remoteTrack *webrtc.TrackRemote, rtpReceiver *webrtc.RTPReceiver) error {
	stream := r.getStream(remoteTrack.StreamID())
	return stream.writeVideoRtp(remoteTrack, r.newTrack)
}

func (r *receiver) getStream(id string) *localStream {
	r.Unlock()
	defer r.Unlock()
	stream, ok := r.streams[id]
	if !ok {
		stream = newLocalStream(id)
		r.streams[id] = stream
	}
	return stream
}

func (r *receiver) stop() {
}
