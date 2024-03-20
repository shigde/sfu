package telemetry

const TracerName = "github.com/shigde/sfu"

type TelemetryConfig struct {
	Enable bool `mapstructure:"enable"`
}

func ValidateTelemetryConfig(_ *TelemetryConfig) error {
	return nil
}
