package rtp

import (
	"sync"

	"github.com/pion/interceptor/pkg/stats"
)

type interceptorMap struct {
	statsLocker *sync.RWMutex
	stats       map[string]stats.Getter
}

func newInterceptorMap() *interceptorMap {
	return &interceptorMap{
		statsLocker: &sync.RWMutex{},
		stats:       make(map[string]stats.Getter),
	}
}

func (i *interceptorMap) setStatsGetter(id string, getter stats.Getter) {
	i.statsLocker.Lock()
	defer i.statsLocker.Unlock()
	i.stats[id] = getter
}
func (i *interceptorMap) getStatsGetter(id string) (stats.Getter, bool) {
	i.statsLocker.RUnlock()
	defer i.statsLocker.RUnlock()
	statsGetter, ok := i.stats[id]
	return statsGetter, ok
}
