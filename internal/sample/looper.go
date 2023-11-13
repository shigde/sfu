package sample

import (
	"fmt"

	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/rtp"
)

type Looper interface {
	sampler
	Codec() webrtc.RTPCodecCapability
}

type videoLooper interface {
	Looper
	ToLayer(quality rtp.VideoQuality) *rtp.SimulcastLayer
}

func CreateVideoLoopers(resolution string, codecFilter string, simulcast bool) ([]videoLooper, error) {
	specs := randomVideoSpecsForCodec(codecFilter)
	numToKeep := 0
	switch resolution {
	case "medium":
		numToKeep = 2
	case "low":
		numToKeep = 1
	default:
		numToKeep = 3
	}
	specs = specs[:numToKeep]
	if !simulcast {
		specs = specs[numToKeep-1:]
	}
	loopers := make([]videoLooper, 0)
	for _, spec := range specs {
		f, err := res.Open(spec.Name())
		if err != nil {
			return nil, err
		}
		defer f.Close()
		if spec.codec == h264Codec {
			looper, err := NewVideoLooperH264(f, spec)
			if err != nil {
				return nil, err
			}
			loopers = append(loopers, looper)
		} else if spec.codec == vp8Codec {
			looper, err := NewVideoLooperVP8(f, spec)
			if err != nil {
				return nil, err
			}
			loopers = append(loopers, looper)
		}
	}
	return loopers, nil
}

func CreateAudioLooper() (*AudioLooperOpus, error) {
	chosenName := audioNames[int(audioIndex.Load())%len(audioNames)]
	audioIndex.Inc()
	f, err := res.Open(fmt.Sprintf("resources/%s.ogg", chosenName))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return NewAudioLooperOpus(f)
}
