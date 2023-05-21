package metric

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type MetricConfig struct {
	Prometheus *PrometheusConfig `mapstructure:"prometheus"`
}

type PrometheusConfig struct {
	Enable   bool   `mapstructure:"enable"`
	Endpoint string `mapstructure:"endpoint"`
}

type Metric struct {
	Endpoint string
	// http metrics
	totalRequests  *prometheus.CounterVec
	responseStatus *prometheus.CounterVec
	httpDuration   *prometheus.HistogramVec
}

func NewMetric(config *MetricConfig) (*Metric, error) {
	Endpoint := config.Prometheus.Endpoint

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

	return &Metric{Endpoint, totalRequests, responseStatus, httpDuration}, nil
}
