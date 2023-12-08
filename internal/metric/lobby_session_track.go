package metric

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var lobbySessionTrackMetric *LobbySessionTrackMetric

type LobbySessionTrackMetric struct {
	runningTracks *prometheus.GaugeVec
}

func RunningTracksInc(session string, track string) {
	if lobbySessionTrackMetric != nil {
		lobbySessionTrackMetric.runningTracks.With(prometheus.Labels{"session": session, "track": track}).Inc()
	}
}
func RunningTracksDec(session string, track string) {
	if lobbySessionTrackMetric != nil {
		lobbySessionTrackMetric.runningTracks.With(prometheus.Labels{"session": session, "track": track}).Dec()
	}
}

func NewLobbySessionTrackMetrics() (*LobbySessionTrackMetric, error) {
	if lobbySessionTrackMetric != nil {
		return nil, errors.New("lobby session track metric already exists")
	}

	lobbySessionTrackMetric = &LobbySessionTrackMetric{
		runningTracks: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shig",
			Name:      "lobby_session_tracks",
			Help:      "running lobby session tracks",
		}, []string{"session", "track"}),
	}
	if err := prometheus.Register(lobbySessionTrackMetric.runningTracks); err != nil {
		return nil, fmt.Errorf("register runningLobby metric: %w", err)
	}

	return lobbySessionTrackMetric, nil
}
