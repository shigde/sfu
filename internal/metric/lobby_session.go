package metric

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var lobbySessions *LobbySessionMetric

type LobbySessionMetric struct {
	RunningSessions prometheus.Gauge
	SessionStates   *prometheus.GaugeVec
}

func RunningSessionsInc() {
	if lobbySessions != nil {
		lobbySessions.RunningSessions.Inc()
	}
}
func RunningSessionsDec() {
	if lobbySessions != nil {
		lobbySessions.RunningSessions.Dec()
	}
}

func NewLobbySessionMetrics() (*LobbySessionMetric, error) {
	m := &LobbySessionMetric{
		RunningSessions: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "shig",
			Name:      "running_lobby_sessions",
			Help:      "Number of currently running lobby sessions.",
		}),
		SessionStates: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shig",
			Name:      "lobby_session_state",
			Help:      "Information about the current lobby session state.",
		}, []string{"session_state"}),
	}
	if err := prometheus.Register(m.RunningSessions); err != nil {
		return nil, fmt.Errorf("register runningSessions metric: %w", err)
	}

	if err := prometheus.Register(m.SessionStates); err != nil {
		return nil, fmt.Errorf("register sessionStates metric: %w", err)
	}
	return m, nil
}
