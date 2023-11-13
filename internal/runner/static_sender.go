package runner

import (
	"context"

	"github.com/shigde/sfu/internal/rtp"
	"github.com/shigde/sfu/internal/static"
	"github.com/shigde/sfu/pkg/client"
)

type StaticSender struct {
	conf                 *rtp.RtpConfig
	audioFile, videoFile string
	spaceId              string
	streamId             string
	bearer               string
}

func (mr *StaticSender) run() {
	media, err := static.NewMediaFile(mr.audioFile, mr.videoFile)
	if err != nil {
		panic(err)
	}

	//conf, err := config.ParseConfig(cli.Config)
	//if err != nil {
	//	panic(err)
	//}

	engine, err := rtp.NewEngine(mr.conf)
	if err != nil {
		panic(err)
	}

	endpoint, err := engine.NewStaticMediaSenderEndpoint(media)
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
