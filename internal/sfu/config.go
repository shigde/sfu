package sfu

import (
	"fmt"

	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/auth"
	"github.com/shigde/sfu/internal/logging"
	"github.com/shigde/sfu/internal/metric"
	"github.com/shigde/sfu/internal/rtp"
	"github.com/shigde/sfu/internal/storage"
	"github.com/shigde/sfu/internal/telemetry"
)

type Config struct {
	*ServerConfig              `mapstructure:"server"`
	*auth.SecurityConfig       `mapstructure:"security"`
	*storage.StorageConfig     `mapstructure:"store"`
	*logging.LogConfig         `mapstructure:"log"`
	*metric.MetricConfig       `mapstructure:"metric"`
	*telemetry.TelemetryConfig `mapstructure:"telemetry"`
	*rtp.RtpConfig             `mapstructure:"rtp"`
	*instance.FederationConfig `mapstructure:"federation"`
}

type Environment struct {
	ServerEnv
	instance.FederationEnv
}

type ServerConfig struct {
	Host  string `mapstructure:"host"`
	Port  int    `mapstructure:"port"`
	HTTPS bool   `mapstructure:"https"`
	Crt   string `mapstructure:"crt"`
	Key   string `mapstructure:"key"`
}

type ServerEnv struct {
	Port int
}

func ValidateServerConfig(config *ServerConfig, env *ServerEnv) error {
	if env.Port >= 1 {
		config.Port = env.Port
	}
	if len(config.Host) < 1 {
		return fmt.Errorf("server.Host should not be empty")
	}
	if config.Port < 1 {
		return fmt.Errorf("server.Port should not be empty")
	}

	if config.HTTPS {
		if len(config.Key) < 1 {
			return fmt.Errorf("server.Key should not be empty")
		}
		if len(config.Crt) < 1 {
			return fmt.Errorf("server.Crt should not be empty")
		}
	}
	return nil
}
