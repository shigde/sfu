package main

import (
	"context"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/config"
	"github.com/shigde/sfu/internal/rtp"
	"github.com/shigde/sfu/internal/runner"
	"github.com/shigde/sfu/pkg/client"
	"github.com/shigde/sfu/pkg/media"
	"github.com/shigde/sfu/pkg/message"
	"golang.org/x/exp/slog"
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

	conf, err := config.ParseConfig(cli.Config, config.ParseEnv())
	if err != nil {
		panic(err)
	}

	engine, err := rtp.NewEngine(conf.RtpConfig)
	if err != nil {
		panic(err)
	}

	signalmesenger := media.NewMessenger()
	handler := media.NewMediaStateEventHandler(signalmesenger)

	signalEndpoint, err := engine.NewSignalConnection(ctx, handler)
	if err != nil {
		panic(err)
	}

	signalOffer, err := signalEndpoint.GetLocalDescription(ctx)
	if err != nil {
		panic(err)
	}

	whipClient := client.NewWhip()
	whipClient.Session.SetBearer(bearer)
	signalAnswer, err := whipClient.GetAnswer(spaceId, streamId, signalOffer)
	if err != nil {
		panic(err)
	}
	println("Answer", signalAnswer.SDP)
	err = signalEndpoint.PeerConnection.SetRemoteDescription(*signalAnswer)
	if err != nil {
		panic(err)
	}

	// Create receive Connection
	//-----
	whepClient := client.NewWhep(client.WithSession(whipClient.Session))
	offer, err := whepClient.GetOffer(spaceId, streamId)
	if err != nil {
		panic(err)
	}

	println("Offer: ##########################################..")
	println("Offer: ", offer.SDP)
	receiveEndpoint, err := engine.NewReceiverConnection(ctx, *offer, &media.EmptyStateHandler{}, rtmpEndpoint)
	if err != nil {
		panic(err)
	}
	observer := newMediaObserver(receiveEndpoint, signalmesenger)
	signalmesenger.Register(observer)

	answer, err := receiveEndpoint.GetLocalDescription(ctx)
	if err != nil {
		panic(err)
	}

	println("Answer: ##########################################..")
	println("Answer: ", answer.SDP)
	err = whepClient.SendAnswer(spaceId, streamId, bearer, answer)
	if err != nil {
		panic(err)
	}

	select {}
}

type mediaObserver struct {
	id        uuid.UUID
	endpoint  *rtp.Connection
	messenger *media.Messenger
}

func (o *mediaObserver) OnMute(mute *message.Mute) {
	//TODO implement me
	panic("implement me")
}

func newMediaObserver(endpoint *rtp.Connection, messenger *media.Messenger) *mediaObserver {
	return &mediaObserver{id: uuid.New(), endpoint: endpoint, messenger: messenger}
}

func (o *mediaObserver) OnOffer(sdp *webrtc.SessionDescription, id uint32, number uint32) {
	slog.Debug("############### Receive an Offer!", "offer")
	println("Offer - x: ##########################################..")
	println("Offer - x: ", &sdp)
	if err := o.endpoint.PeerConnection.SetRemoteDescription(*sdp); err != nil {
		slog.Error("set remote description", "err", err)
		return
	}

	o.endpoint.GatherComplete = webrtc.GatheringCompletePromise(o.endpoint.PeerConnection)
	answer, err := o.endpoint.PeerConnection.CreateAnswer(nil)
	if err != nil {
		slog.Error("create answer", "err", err)
		return
	}

	if err = o.endpoint.PeerConnection.SetLocalDescription(answer); err != nil {
		slog.Error("set answer", "err", err)
		return
	}

	ldc, err := o.endpoint.GetLocalDescription(context.Background())
	if err != nil {
		slog.Error("get local dc", "err", err)
		return
	}

	_, err = o.messenger.SendSDP(ldc, id, number)
	if err != nil {
		slog.Error("send local dc", "err", err)

	}
	return
}
func (o *mediaObserver) OnAnswer(sdp *webrtc.SessionDescription, id uint32, number uint32) {
}
func (o *mediaObserver) GetId() uuid.UUID {
	return o.id
}
