package rtp

import "github.com/pion/webrtc/v3"

type sender struct {
	stream *receiverStream
}

type senderStream struct {
	audioTrack *webrtc.TrackLocalStaticRTP
	// @TODO: Change this, because maybe this should be a list of video tracks
	videoTrack *webrtc.TrackLocalStaticRTP
}
