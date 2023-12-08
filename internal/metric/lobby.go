package metric

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var lobbyMetric *LobbyMetric

type LobbyMetric struct {
	runningLobby *prometheus.GaugeVec
}

func RunningLobbyInc(stream string, lobby string) {
	if lobbyMetric != nil {
		lobbyMetric.runningLobby.With(prometheus.Labels{"stream": stream, "lobby": lobby}).Inc()
	}
}
func RunningLobbyDec(stream string, lobby string) {
	if lobbyMetric != nil {
		lobbyMetric.runningLobby.With(prometheus.Labels{"stream": stream, "lobby": lobby}).Dec()
	}
}

func NewLobbyMetrics() (*LobbyMetric, error) {
	if lobbyMetric != nil {
		return nil, errors.New("lobby metric already exists")
	}

	lobbyMetric = &LobbyMetric{
		runningLobby: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shig",
			Name:      "lobby",
			Help:      "running lobbies",
		}, []string{"stream", "lobby"}),
	}
	if err := prometheus.Register(lobbyMetric.runningLobby); err != nil {
		return nil, fmt.Errorf("register runningLobby metric: %w", err)
	}

	return lobbyMetric, nil
}
