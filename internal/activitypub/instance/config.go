package instance

import (
	"database/sql"
	"fmt"
	"net/url"
	"time"
)

type FederationConfig struct {
	Https            bool              `mapstructure:"https"`
	Enable           bool              `mapstructure:"enable"`
	Domain           string            `mapstructure:"domain"`
	Release          string            `mapstructure:"release"`
	InstanceUsername string            `mapstructure:"instanceUsername"`
	ServerName       string            `mapstructure:"serverName"`
	IsPrivate        bool              `mapstructure:"private"`
	RegisterToken    string            `mapstructure:"registerToken"`
	TrustedInstances []TrustedInstance `mapstructure:"trustedInstance"`
	InstanceUrl      *url.URL
	ServerInitTime   sql.NullTime
}

type TrustedInstance struct {
	Actor string `mapstructure:"actor"`
	Name  string `mapstructure:"name"`
}

type FederationEnv struct {
	Domain        string
	RegisterToken string
}

func ValidateFederationConfig(config *FederationConfig, env *FederationEnv) error {
	if !config.Enable {
		return nil
	}

	// Env overrides config file
	if len(env.Domain) > 0 {
		config.Domain = env.Domain
	}

	// Env overrides config file
	if len(env.RegisterToken) > 0 {
		config.RegisterToken = env.RegisterToken
	}

	if len(config.InstanceUsername) < 1 {
		config.InstanceUsername = "shig"
	}

	if len(config.ServerName) < 1 {
		config.ServerName = "shig"
	}

	if len(config.RegisterToken) < 1 {
		return fmt.Errorf("Federation is enabled but register token is not set properly.")
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

	if len(config.TrustedInstances) != 0 {
		for n, inst := range config.TrustedInstances {
			if err := validateTrustedInstances(inst, n); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateTrustedInstances(inst TrustedInstance, n int) error {
	if _, err := url.Parse(inst.Actor); err != nil {
		return fmt.Errorf("federation.trustedInstances[]{actor=} has to be set for an instance, entry %d", n)
	}
	if len(inst.Name) < 1 {
		return fmt.Errorf("federation.trustedInstances[]{name=} has to be set for an instance, entry %d", n)
	}

	return nil
}
