package activitypub

import (
	"fmt"
)

type FederationConfig struct {
	Enable           bool   `mapstructure:"enable"`
	InstanceHostname string `mapstructure:"instanceHostname"`
}

func ValidateFederationConfig(config *FederationConfig) error {
	if !config.Enable {
		return nil
	}

	if len(config.InstanceHostname) < 3 {
		return fmt.Errorf("Federation is enabled but instance host name is not set properly.")
	}
	return nil
}
