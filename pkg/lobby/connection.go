package lobby

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/pion/webrtc/v3"
)

type Connection struct {
	peerConnection *webrtc.PeerConnection
}

func newConnection() (*Connection, error) {
	api := webrtc.NewAPI()
	// @TODO Check engine setup
	peerConnection, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return nil, fmt.Errorf("creating peer connection: %w", err)
	}

	// On Track benötige ich dann nur wenn ich einen publisher habe
	// Ich erzeuge lokale Tracks und kopiere die Remote tracks dort rein
	// Warum kann ich nicht die remote tracks benutzen?
	// create Local Tracks!!!!!!!
	peerConnection.OnTrack(func(remoteTrack *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		var localTrack *webrtc.TrackLocalStaticRTP
		if strings.HasPrefix(remoteTrack.Codec().RTPCodecCapability.MimeType, "audio") {
			//localTrack = audioTrack
		} else {
			//localTrack = videoTrack
		}

		//nolint:all
		rtpBuf := make([]byte, 1500)
		for {
			rtpRead, _, readErr := remoteTrack.Read(rtpBuf)
			switch {
			case errors.Is(readErr, io.EOF):
				return
			case readErr != nil:
				log.Println(readErr)
				return
			}

			if _, writeErr := localTrack.Write(rtpBuf[:rtpRead]); writeErr != nil && !errors.Is(writeErr, io.ErrClosedPipe) {
				log.Println(writeErr)
				return
			}
		}
	})

	//	peerConnection, err := api.NewPeerConnection(webrtc.Configuration{})
	//	if err != nil {
	//		logHTTPError(w, err.Error(), http.StatusBadRequest)
	//		return
	//	}
	//
	//	audioTrack, videoTrack, err := getTracksForStream(streamKey)
	//	if err != nil {
	//		logHTTPError(w, err.Error(), http.StatusBadRequest)
	//		return
	//	}
	//
	//	peerConnection.OnTrack(func(remoteTrack *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
	//		var localTrack *webrtc.TrackLocalStaticRTP
	//		if strings.HasPrefix(remoteTrack.Codec().RTPCodecCapability.MimeType, "audio") {
	//			localTrack = audioTrack
	//		} else {
	//			localTrack = videoTrack
	//		}
	//
	//		//nolint:all
	//		rtpBuf := make([]byte, 1500)
	//		for {
	//			rtpRead, _, readErr := remoteTrack.Read(rtpBuf)
	//			switch {
	//			case errors.Is(readErr, io.EOF):
	//				return
	//			case readErr != nil:
	//				log.Println(readErr)
	//				return
	//			}
	//
	//			if _, writeErr := localTrack.Write(rtpBuf[:rtpRead]); writeErr != nil && !errors.Is(writeErr, io.ErrClosedPipe) {
	//				log.Println(writeErr)
	//				return
	//			}
	//		}
	//	})
	//
	//	peerConnection.OnICEConnectionStateChange(func(i webrtc.ICEConnectionState) {
	//		if i == webrtc.ICEConnectionStateFailed {

	return &Connection{peerConnection}, nil

}
