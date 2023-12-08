package metric

import "fmt"

type MetricConfig struct {
	Prometheus *PrometheusConfig `mapstructure:"prometheus"`
}

type PrometheusConfig struct {
	Enable   bool   `mapstructure:"enable"`
	Endpoint string `mapstructure:"endpoint"`
	Port     int    `mapstructure:"port"`
}

func ValidateMetricConfig(config *MetricConfig) error {
	if len(config.Prometheus.Endpoint) == 0 {
		return fmt.Errorf("metric.prometheus.endpoint should not be empty")
	}

	if config.Prometheus.Port < 1 {
		return fmt.Errorf("metric.prometheus.port should not be empty")
	}

	return nil
}
