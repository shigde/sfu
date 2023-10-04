package telemetry

type TelemetryConfig struct {
	Enable bool `mapstructure:"enable"`
}

func ValidateTelemetryConfig(_ *TelemetryConfig) error {
	return nil
}
