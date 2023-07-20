package rtp

import (
	"errors"
	"io"
	"log"

	"github.com/pion/webrtc/v3"
)

const rtpBufferSize = 1500

type receiver struct {
	stream *receiverStream
}

func newReceiver() *receiver {
	stream := newReceiverStream()
	return &receiver{stream}
}

func (w *receiver) audioWrite(remoteTrack *webrtc.TrackRemote) {
	rtpBuf := make([]byte, rtpBufferSize)
	for {
		_, _, err := remoteTrack.Read(rtpBuf)
		switch {
		case errors.Is(err, io.EOF):
			return
		case err != nil:
			log.Println(err)
			return
		}

		//if _, writeErr := w.stream.audioTrack.Write(rtpBuf[:rtpRead]); writeErr != nil && !errors.Is(writeErr, io.ErrClosedPipe) {
		//	log.Println(writeErr)
		//	return
		//}
	}
}

func (w *receiver) videoWrite(remoteTrack *webrtc.TrackRemote) {
	rtpBuf := make([]byte, rtpBufferSize)
	for {
		_, _, err := remoteTrack.Read(rtpBuf)
		switch {
		case errors.Is(err, io.EOF):
			return
		case err != nil:
			log.Println(err)
			return
		}

		//if _, writeErr := w.stream.audioTrack.Write(rtpBuf[:rtpRead]); writeErr != nil && !errors.Is(writeErr, io.ErrClosedPipe) {
		//	log.Println(writeErr)
		//	return
		//}
	}
}

func (w *receiver) stop() {
}

// @TODO: At the moment I'm not sure if I really need an TrackLocalStaticRTP here or if I can't use the remote track directly. so I make a distinction here.
type receiverStream struct {
	audioTrack *webrtc.TrackLocalStaticRTP
	// @TODO: Change this, because maybe this should be a list of video tracks
	videoTrack *webrtc.TrackLocalStaticRTP
}

func newReceiverStream() *receiverStream {
	return &receiverStream{}
}
