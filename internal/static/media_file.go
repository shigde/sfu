package static

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/ivfreader"
	"github.com/pion/webrtc/v3/pkg/media/oggreader"
)

type MediaFile struct {
	streamId                     string
	audioFileName, videoFileName string
	AudioTrack, VideoTrack       *webrtc.TrackLocalStaticSample
	oggPageDuration              time.Duration
}

func NewMediaFile(audioFileName string, videoFileName string) (*MediaFile, error) {

	// Assert that we have an audio or video file
	if _, err := os.Stat(videoFileName); err != nil {
		switch {
		case os.IsNotExist(err):
			return nil, fmt.Errorf("video file not exists: %w", err)
		default:
			return nil, fmt.Errorf("reading video file: %w", err)
		}
	}

	if _, err := os.Stat(audioFileName); err != nil {
		switch {
		case os.IsNotExist(err):
			return nil, fmt.Errorf("audio file not exists: %w", err)
		default:
			return nil, fmt.Errorf("reading audio file: %w", err)
		}
	}

	m := &MediaFile{
		streamId:        uuid.NewString(),
		audioFileName:   audioFileName,
		videoFileName:   videoFileName,
		oggPageDuration: time.Millisecond * 20,
	}

	videoTrack, err := m.GetVideoTrack()
	if err != nil {
		return nil, fmt.Errorf("creating video track: %w", err)
	}

	audioTrack, err := m.GetAudioTrack()
	if err != nil {
		return nil, fmt.Errorf("creating audio track: %w", err)
	}
	m.VideoTrack = videoTrack
	m.AudioTrack = audioTrack

	return m, nil
}

func (m *MediaFile) GetVideoTrack() (*webrtc.TrackLocalStaticSample, error) {
	file, err := os.Open(m.videoFileName)
	if err != nil {
		return nil, fmt.Errorf("reading video file: %w", err)
	}

	_, header, err := ivfreader.NewWith(file)
	if err != nil {
		return nil, fmt.Errorf("open  video  ivfreader: %w", err)
	}

	// Determine video codec
	var trackCodec string
	switch header.FourCC {
	// case "AV01":
	// trackCodec = webrtc.MimeTypeAV1
	case "VP90":
		trackCodec = webrtc.MimeTypeVP9
	case "VP80":
		trackCodec = webrtc.MimeTypeVP8
	default:
		return nil, fmt.Errorf("handle FourCC %s: %w", header.FourCC, err)
	}

	// Create a video track
	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: trackCodec}, uuid.NewString(), m.streamId)
	if err != nil {
		return nil, fmt.Errorf("open  video  ivfreader: %w", err)
	}

	return videoTrack, nil
}

func (m *MediaFile) GetAudioTrack() (*webrtc.TrackLocalStaticSample, error) {
	// Create a audio track
	audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, uuid.NewString(), m.streamId)
	if err != nil {
		panic(err)
	}

	return audioTrack, nil
}

func (m *MediaFile) PlayVideo(iceConnectedCtx context.Context, rtpSender *webrtc.RTPSender) {
	// Read incoming RTCP packets
	// Before these packets are returned they are processed by interceptors. For things
	// like NACK this needs to be called.
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	go func() {
		// Open a IVF file and start reading using our IVFReader
		file, ivfErr := os.Open(m.videoFileName)
		if ivfErr != nil {
			panic(ivfErr)
		}

		ivf, header, ivfErr := ivfreader.NewWith(file)
		if ivfErr != nil {
			panic(ivfErr)
		}

		// Wait for connection established
		<-iceConnectedCtx.Done()

		// Send our video file frame at a time. Pace our sending so we send it at the same speed it should be played back as.
		// This isn't required since the video is timestamped, but we will such much higher loss if we send all at once.
		//
		// It is important to use a time.Ticker instead of time.Sleep because
		// * avoids accumulating skew, just calling time.Sleep didn't compensate for the time spent parsing the data
		// * works around latency issues with Sleep (see https://github.com/golang/go/issues/44343)
		ticker := time.NewTicker(time.Millisecond * time.Duration((float32(header.TimebaseNumerator)/float32(header.TimebaseDenominator))*1000))
		for ; true; <-ticker.C {
			frame, _, ivfErr := ivf.ParseNextFrame()
			if errors.Is(ivfErr, io.EOF) {
				fmt.Printf("All video frames parsed and sent")
				os.Exit(0)
			}

			if ivfErr != nil {
				panic(ivfErr)
			}

			if ivfErr = m.VideoTrack.WriteSample(media.Sample{Data: frame, Duration: time.Second}); ivfErr != nil {
				panic(ivfErr)
			}
		}
	}()
}

func (m *MediaFile) PlayAudio(iceConnectedCtx context.Context, rtpSender *webrtc.RTPSender) {
	// Read incoming RTCP packets
	// Before these packets are returned they are processed by interceptors. For things
	// like NACK this needs to be called.
	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	go func() {
		// Open a OGG file and start reading using our OGGReader
		file, err := os.Open(m.audioFileName)
		if err != nil {
			panic(err)
		}

		// Open on oggfile in non-checksum mode.
		ogg, _, err := oggreader.NewWith(file)
		if err != nil {
			panic(err)
		}

		// Wait for connection established
		<-iceConnectedCtx.Done()

		// Keep track of last granule, the difference is the amount of samples in the buffer
		var lastGranule uint64

		// It is important to use a time.Ticker instead of time.Sleep because
		// * avoids accumulating skew, just calling time.Sleep didn't compensate for the time spent parsing the data
		// * works around latency issues with Sleep (see https://github.com/golang/go/issues/44343)
		ticker := time.NewTicker(m.oggPageDuration)
		for ; true; <-ticker.C {
			pageData, pageHeader, oggErr := ogg.ParseNextPage()
			if errors.Is(oggErr, io.EOF) {
				fmt.Printf("All audio pages parsed and sent")
				os.Exit(0)
			}

			if oggErr != nil {
				panic(oggErr)
			}

			// The amount of samples is the difference between the last and current timestamp
			sampleCount := float64(pageHeader.GranulePosition - lastGranule)
			lastGranule = pageHeader.GranulePosition
			sampleDuration := time.Duration((sampleCount/48000)*1000) * time.Millisecond

			if oggErr = m.AudioTrack.WriteSample(media.Sample{Data: pageData, Duration: sampleDuration}); oggErr != nil {
				panic(oggErr)
			}
		}
	}()
}

func (m *MediaFile) Stop() {

}
