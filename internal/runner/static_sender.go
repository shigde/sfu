package runner

import (
	"context"
	"fmt"

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

func (mr *StaticSender) run(done chan struct{}) error {
	localTracks := make([]webrtc.TrackLocal, 0, 2)

	videoTrack, err := sample.NewLocalFileLooperTrack(mr.videoFile)
	if err != nil {
		return fmt.Errorf("creating video track: %w", err)
	}
	localTracks = append(localTracks, videoTrack)

	audioTrack, err := sample.NewLocalFileLooperTrack(mr.audioFile)
	if err != nil {
		return fmt.Errorf("creating audio track: %w", err)
	}
	localTracks = append(localTracks, audioTrack)

	engine, err := rtp.NewEngine(mr.conf)
	if err != nil {
		return fmt.Errorf("setup webrtc engine: %w", err)
	}

	endpoint, err := rtp.NewLocalStaticSenderEndpoint(engine, localTracks)
	if err != nil {
		return fmt.Errorf("building new webrtc endpoint: %w", err)
	}

	offer, err := endpoint.GetLocalDescription(context.Background())
	if err != nil {
		return fmt.Errorf("creating local offer: %w", err)
	}

	whipClient := client.NewWhip()
	answer, err := whipClient.GetAnswer(mr.spaceId, mr.streamId, mr.bearer, offer)
	if err != nil {
		return fmt.Errorf("getting answer from whip endpoint: %w", err)
	}
	println("Answer", answer.SDP)
	err = endpoint.SetAnswer(answer)
	if err != nil {
		return fmt.Errorf("creating setting answer: %w", err)
	}

	select {
	case <-done:
		return nil
	}
}
