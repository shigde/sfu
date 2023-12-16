package metric

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var lobbySessions *LobbySessionMetric

type LobbySessionMetric struct {
	runningSessions *prometheus.GaugeVec
}

func RunningSessionsInc(lobby string, session string) {
	if lobbySessions != nil {
		lobbySessions.runningSessions.With(prometheus.Labels{"lobby": lobby, "session": session}).Inc()
	}
}
func RunningSessionsDec(lobby string, session string) {
	if lobbySessions != nil {
		lobbySessions.runningSessions.Delete(prometheus.Labels{"lobby": lobby, "session": session})
	}
}

func NewLobbySessionMetrics() (*LobbySessionMetric, error) {
	if lobbySessions != nil {
		return nil, errors.New("lobby session metric already exists")
	}
	lobbySessions = &LobbySessionMetric{
		runningSessions: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shig",
			Name:      "lobby_session",
			Help:      "running lobby sessions",
		}, []string{"lobby", "session"}),
	}
	if err := prometheus.Register(lobbySessions.runningSessions); err != nil {
		return nil, fmt.Errorf("register runningSessions metric: %w", err)
	}
	return lobbySessions, nil
}
