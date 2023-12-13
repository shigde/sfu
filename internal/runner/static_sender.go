package runner

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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
	session              *client.Session
	isMainStream         bool
}

func NewStaticSender(conf *rtp.RtpConfig, audioFile string, videoFile string, spaceId string, streamId string, session *client.Session, isMainStream bool) *StaticSender {
	return &StaticSender{
		conf,
		audioFile,
		videoFile,
		spaceId,
		streamId,
		session,
		isMainStream,
	}
}

func (mr *StaticSender) Run(ctx context.Context, onEstablished chan<- struct{}) error {
	localTracks := make([]webrtc.TrackLocal, 0, 2)

	streamID := uuid.NewString()
	videoTrack, err := sample.NewLocalFileLooperTrack(mr.videoFile, sample.WithStreamID(streamID))
	if err != nil {
		return fmt.Errorf("creating video track: %w", err)
	}
	localTracks = append(localTracks, videoTrack)

	audioTrack, err := sample.NewLocalFileLooperTrack(mr.audioFile, sample.WithStreamID(streamID))
	if err != nil {
		return fmt.Errorf("creating audio track: %w", err)
	}
	localTracks = append(localTracks, audioTrack)

	engine, err := rtp.NewEngine(mr.conf)
	if err != nil {
		return fmt.Errorf("setup webrtc engine: %w", err)
	}

	withOnEstablished := rtp.EndpointWithOnEstablished(func() {
		select {
		case <-ctx.Done():
		default:
			onEstablished <- struct{}{}
		}
	})

	endpoint, err := rtp.EstablishStaticIngressEndpoint(engine, localTracks, withOnEstablished)
	if err != nil {
		return fmt.Errorf("building new webrtc endpoint: %w", err)
	}

	offer, err := endpoint.GetLocalDescription(ctx)
	if err != nil {
		return fmt.Errorf("creating local offer: %w", err)
	}

	if mr.isMainStream {
		if offer, err = rtp.MarkStreamAsMain(offer, streamID); err != nil {
			return fmt.Errorf("marking straem as main stream: %w", err)
		}
	}

	whipClient := client.NewWhip(client.WithSession(mr.session))
	answer, err := whipClient.GetAnswer(mr.spaceId, mr.streamId, offer)
	if err != nil {
		return fmt.Errorf("getting answer from whip endpoint: %w", err)
	}

	err = endpoint.SetAnswer(answer)
	if err != nil {
		return fmt.Errorf("creating setting answer: %w", err)
	}

	select {
	case <-ctx.Done():
		return nil
	}
}
