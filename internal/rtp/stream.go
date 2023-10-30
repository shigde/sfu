package rtp

import (
	"context"

	"github.com/pion/webrtc/v3"
)

type Stream interface {
	writeAudioRtp(ctx context.Context, track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) error
	writeVideoRtp(ctx context.Context, track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) error
	getVideoTrack() *webrtc.TrackLocalStaticRTP
	getAudioTrack() *webrtc.TrackLocalStaticRTP
	getLiveVideoTrack() *LiveTrackStaticRTP
	getLiveAudioTrack() *LiveTrackStaticRTP
	getKind() TrackInfoKind
}
