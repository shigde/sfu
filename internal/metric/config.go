package metric

type MetricConfig struct {
	Prometheus *PrometheusConfig `mapstructure:"prometheus"`
}

type PrometheusConfig struct {
	Enable   bool   `mapstructure:"enable"`
	Endpoint string `mapstructure:"endpoint"`
}
