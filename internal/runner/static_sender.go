package runner

import (
	"context"

	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/rtp"
	"github.com/shigde/sfu/internal/sample"
	"github.com/shigde/sfu/pkg/client"
)

type StaticSender struct {
	conf                 *rtp.RtpConfig
	audioFile, videoFile string
	spaceId              string
	streamId             string
	bearer               string
}

func (mr *StaticSender) run() error {
	localTracks := make([]webrtc.TrackLocal, 0, 2)

	videoTrack, err := sample.NewLocalFileLooperTrack(mr.videoFile)
	if err != nil {
		return err
	}
	localTracks = append(localTracks, videoTrack)

	audioTrack, err := sample.NewLocalFileLooperTrack(mr.audioFile)
	if err != nil {
		return err
	}
	localTracks = append(localTracks, audioTrack)

	engine, err := rtp.NewEngine(mr.conf)
	if err != nil {
		panic(err)
	}

	endpoint, err := engine.NewLocalStaticSenderEndpoint(localTracks)
	if err != nil {
		panic(err)
	}

	offer, err := endpoint.GetLocalDescription(context.Background())
	if err != nil {
		panic(err)
	}

	whipClient := client.NewWhip()
	answer, err := whipClient.GetAnswer(mr.spaceId, mr.streamId, mr.bearer, offer)
	if err != nil {
		panic(err)
	}
	println("Answer", answer.SDP)
	err = endpoint.SetAnswer(answer)
	if err != nil {
		panic(err)
	}

	select {}
}
