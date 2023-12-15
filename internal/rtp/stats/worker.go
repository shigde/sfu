package stats

import (
	"time"

	"github.com/pion/interceptor/pkg/stats"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/metric"
)

func worker(labels metric.Labels, ssrc webrtc.SSRC, statsGetter stats.Getter, cancel <-chan struct{}) {
	for {
		select {
		case <-cancel:
			metric.CleanTrackStats(labels)
			return
		case <-time.After(5 * time.Second):
			statsRep := statsGetter.Get(uint32(ssrc))
			metric.RecordTrackStats(labels, statsRep)
		}
	}
}
