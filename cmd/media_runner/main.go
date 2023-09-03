package main

import (
	"context"
	"fmt"

	"github.com/shigde/sfu/internal/config"
	"github.com/shigde/sfu/internal/rtp"
	"github.com/shigde/sfu/internal/runner"
	"github.com/shigde/sfu/internal/static"
	"github.com/shigde/sfu/pkg/client"
)

const (
	spaceId  = "123"
	streamId = "value"
	bearer   = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwiaWF0IjoxNTE2MjM5MDIyLCJ1dWlkIjoiYTY0MzY1ZGItMTc0ZC00ZDExLThjYjEtZWIyYTM2MzlmZmU2In0._xbasA_1ljeszeWdqYqp96EWvJIbCnYOTOFxKgcd7vM"
)

func main() {
	// 1. load movie file
	// 2. transcode
	// 3. create endpoint with tracks and offer
	// 2. send offer to sfu over whip
	// 4. Get answer and set answer to endpoint

	cli := runner.NewCli()
	cli.Parse()

	media, err := static.NewMediaFile(cli.AudioFileName, cli.VideoFileName)
	if err != nil {
		panic(err)
	}

	conf, err := config.ParseConfig(cli.Config)
	if err != nil {
		panic(fmt.Errorf("parsing config: %w", err))
		return
	}

	engine, err := rtp.NewEngine(conf.RtpConfig)
	if err != nil {
		panic(err)
	}

	endpoint, err := engine.NewMediaSenderEndpoint(media)
	if err != nil {
		panic(err)
	}

	offer, err := endpoint.GetLocalDescription(context.Background())
	if err != nil {
		panic(err)
	}

	whipClient := client.NewWhip()
	answer, err := whipClient.GetAnswer(spaceId, streamId, bearer, offer)
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
