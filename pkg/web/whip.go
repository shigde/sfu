package web

import "net/http"

// import (
//
//	"fmt"
//	"io"
//	"io/ioutil"
//	"log"
//	"net/http"
//	"strings"
//
//	"github.com/pion/webrtc/v3"
//
// )
func WhipHandler(w http.ResponseWriter, r *http.Request) {

}

//	streamKey := r.Header.Get("Authorization")
//	if streamKey == "" {
//		logHTTPError(w, "Authorization was not set", http.StatusBadRequest)
//		return
//	}
//
//	offer, err := ioutil.ReadAll(r.Body)
//	if err != nil {
//		logHTTPError(w, err.Error(), http.StatusBadRequest)
//		return
//	}
//
//	//nolint:all
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
//			if err := peerConnection.Close(); err != nil {
//				log.Println(err)
//				return
//			}
//		}
//	})
//
//	if err := peerConnection.SetRemoteDescription(webrtc.SessionDescription{
//		SDP:  string(offer),
//		Type: webrtc.SDPTypeOffer,
//	}); err != nil {
//		logHTTPError(w, err.Error(), http.StatusBadRequest)
//		return
//	}
//
//	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
//	answer, err := peerConnection.CreateAnswer(nil)
//
//	if err != nil {
//		logHTTPError(w, err.Error(), http.StatusBadRequest)
//		return
//	} else if err = peerConnection.SetLocalDescription(answer); err != nil {
//		logHTTPError(w, err.Error(), http.StatusBadRequest)
//		return
//	}
//
//	<-gatherComplete
//
//	w.WriteHeader(http.StatusCreated)
//	fmt.Fprint(w, peerConnection.LocalDescription().SDP)
//}
