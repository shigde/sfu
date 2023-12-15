package stats

import (
	"errors"
	"sync"

	"github.com/pion/interceptor/pkg/stats"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/metric"
)

var ErrTrackAlreadyRegistered = errors.New("track source already registered")

type Registry struct {
	sync.RWMutex
	session     string
	statsList   map[webrtc.SSRC]chan struct{}
	statsGetter stats.Getter
}

func NewRegistry(session string, getter stats.Getter) *Registry {
	return &Registry{
		RWMutex:     sync.RWMutex{},
		session:     session,
		statsList:   make(map[webrtc.SSRC]chan struct{}),
		statsGetter: getter,
	}
}

func (r *Registry) StartWorker(labels metric.Labels, ssrc webrtc.SSRC) error {
	r.Lock()
	if _, ok := r.statsList[ssrc]; ok {
		return ErrTrackAlreadyRegistered
	}
	cancel := make(chan struct{})
	r.statsList[ssrc] = cancel
	r.Unlock()

	labels[metric.Session] = r.session
	labels[metric.SSRC] = SSRCtoString(ssrc)

	go worker(labels, ssrc, r.statsGetter, cancel)

	return nil
}

func (r *Registry) StopWorker(ssrc webrtc.SSRC) {
	r.Lock()
	defer r.Unlock()
	if cancel, ok := r.statsList[ssrc]; ok {
		delete(r.statsList, ssrc)
		close(cancel)
	}
}

func (r *Registry) StopAllWorker() {
	r.Lock()
	defer r.Unlock()
	for _, cancel := range r.statsList {
		close(cancel)
	}
	r.statsList = make(map[webrtc.SSRC]chan struct{})
}
