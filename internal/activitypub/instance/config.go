package instance

import (
	"database/sql"
	"fmt"
	"net/url"
	"time"
)

type FederationConfig struct {
	Https            bool   `mapstructure:"https"`
	Enable           bool   `mapstructure:"enable"`
	Domain           string `mapstructure:"domain"`
	Release          string `mapstructure:"release"`
	InstanceUsername string `mapstructure:"instanceUsername"`
	ServerName       string `mapstructure:"serverName"`
	IsPrivate        bool   `mapstructure:"private"`
	InstanceUrl      *url.URL
	ServerInitTime   sql.NullTime
}

func ValidateFederationConfig(config *FederationConfig) error {
	if !config.Enable {
		return nil
	}

	if len(config.InstanceUsername) < 1 {
		config.InstanceUsername = "shig"
	}

	if len(config.ServerName) < 1 {
		config.ServerName = "shig"
	}

	if len(config.Domain) < 3 {
		return fmt.Errorf("Federation is enabled but domain is not set properly.")
	}

	if len(config.Release) < 1 {
		return fmt.Errorf("Federation is enabled but release vesion is not set properly.")
	}
	protocol := "http"
	if config.Https {
		protocol = protocol + "s"
	}
	instanceURL, err := url.Parse(fmt.Sprintf("%s://%s", protocol, config.Domain))
	if err != nil {
		return fmt.Errorf("parsing instance url: %w", err)
	}
	config.InstanceUrl = instanceURL

	config.ServerInitTime = sql.NullTime{
		Time:  time.Time{},
		Valid: true,
	}

	return nil
}
