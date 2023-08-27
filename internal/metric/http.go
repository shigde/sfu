package metric

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type HttpMetric struct {
	// http metrics
	totalRequests  *prometheus.CounterVec
	responseStatus *prometheus.CounterVec
	httpDuration   *prometheus.HistogramVec
}

func NewHttpMetric() (*HttpMetric, error) {
	totalRequests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Number of get requests.",
		},
		[]string{"path"},
	)

	responseStatus := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "response_status",
			Help: "Status of HTTP response",
		},
		[]string{"status"},
	)

	httpDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_response_time_seconds",
			Help: "Duration of HTTP requests.",
		}, []string{"path"})

	if err := prometheus.Register(totalRequests); err != nil {
		return nil, fmt.Errorf("register totalRequest metric: %w", err)
	}
	if err := prometheus.Register(responseStatus); err != nil {
		return nil, fmt.Errorf("register responseStatus metric: %w", err)
	}
	if err := prometheus.Register(httpDuration); err != nil {
		return nil, fmt.Errorf("register httpDuration metric: %w", err)
	}

	return &HttpMetric{totalRequests, responseStatus, httpDuration}, nil
}
