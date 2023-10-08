package main

import (
	"context"

	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/config"
	"github.com/shigde/sfu/internal/rtp"
	"github.com/shigde/sfu/internal/runner"
	"github.com/shigde/sfu/pkg/client"
)

const (
	spaceId  = "live_stream_channel@localhost:9000"
	streamId = "33ccbfbc-d07c-4ed4-abd6-077ba9cb9e65"
	bearer   = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwiaWF0IjoxNTE2MjM5MDIyLCJ1dWlkIjoiYTY0MzY1ZGItMTc0ZC00ZDExLThjYjEtZWIyYTM2MzlmZmU2In0._xbasA_1ljeszeWdqYqp96EWvJIbCnYOTOFxKgcd7vM"

	rtmpEndpoint = "rtmp://127.0.0.1:1935/live/15d2f10a-ba68-46c7-8755-52e9320cbd47"
)

func main() {
	ctx := context.Background()
	cli := runner.NewCli()
	cli.Parse()

	conf, err := config.ParseConfig(cli.Config)
	if err != nil {
		panic(err)
	}

	engine, err := rtp.NewEngine(conf.RtpConfig)
	if err != nil {
		panic(err)
	}

	signalEndpoint, err := engine.NewStaticSignalEndpoint(ctx, newMediaStateEventHandler())
	if err != nil {
		panic(err)
	}

	signalOffer, err := signalEndpoint.GetLocalDescription(ctx)
	if err != nil {
		panic(err)
	}

	whipClient := client.NewWhip()
	signalAnswer, err := whipClient.GetAnswer(spaceId, streamId, bearer, signalOffer)
	if err != nil {
		panic(err)
	}
	println("Answer", signalAnswer.SDP)
	err = signalEndpoint.SetAnswer(signalAnswer)
	if err != nil {
		panic(err)
	}

	// Create receive Connection
	//-----
	whepClient := client.NewWhep(whipClient.Session, whipClient.CsrfToken)
	offer, err := whepClient.GetOffer(spaceId, streamId, bearer)
	if err != nil {
		panic(err)
	}

	receiveEndpoint, err := engine.NewStaticReceiverEndpoint(ctx, *offer, newMediaStateEventHandler(), rtmpEndpoint)
	if err != nil {
		panic(err)
	}

	answer, err := receiveEndpoint.GetLocalDescription(ctx)
	if err != nil {
		panic(err)
	}

	err = whepClient.SendAnswer(spaceId, streamId, bearer, answer)
	if err != nil {
		panic(err)
	}

	select {}
}

type mediaStateEventHandler struct {
}

func newMediaStateEventHandler() *mediaStateEventHandler {
	return &mediaStateEventHandler{}
}

func (h *mediaStateEventHandler) OnConnectionStateChange(state webrtc.ICEConnectionState) {
}
func (h *mediaStateEventHandler) OnNegotiationNeeded(offer webrtc.SessionDescription) {}
func (h *mediaStateEventHandler) OnChannel(dc *webrtc.DataChannel)                    {}
