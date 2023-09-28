package instance

import (
	"fmt"
)

type FederationConfig struct {
	Https   bool   `mapstructure:"https"`
	Enable  bool   `mapstructure:"enable"`
	Domain  string `mapstructure:"domain"`
	Release string `mapstructure:"release"`
}

func ValidateFederationConfig(config *FederationConfig) error {
	if !config.Enable {
		return nil
	}

	if len(config.Domain) < 3 {
		return fmt.Errorf("Federation is enabled but domain is not set properly.")
	}

	if len(config.Release) < 1 {
		return fmt.Errorf("Federation is enabled but release vesion is not set properly.")
	}
	return nil
}
