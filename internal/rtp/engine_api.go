package rtp

import (
	"github.com/pion/interceptor/pkg/stats"
	"github.com/pion/webrtc/v3"
)

type engineApi struct {
	*webrtc.API
	onStatsGetter func(getter stats.Getter)
}

type engineApiOption func(enginApi *engineApi)

func withOnStatsGetter(onStatsGetter func(getter stats.Getter)) func(api *engineApi) {
	return func(api *engineApi) {
		api.onStatsGetter = onStatsGetter
	}
}
