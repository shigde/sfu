package rtp

import (
	"fmt"
)

type RtpConfig struct {
	IceServer []IceServer `mapstructure:"iceServer"`
}

type IceServer struct {
	Urls           []string `mapstructure:"urls"`
	Username       string   `mapstructure:"username"`
	Credential     string   `mapstructure:"credential"`
	CredentialType string   `mapstructure:"credentialType"`
}

func ValidateRtpConfig(config *RtpConfig) error {
	if len(config.IceServer) != 0 {
		for n, server := range config.IceServer {
			if err := validateConfigIceServer(server, n); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateConfigIceServer(iceServer IceServer, n int) error {
	if iceServer.Urls == nil || len(iceServer.Urls) == 0 {
		return fmt.Errorf("urls has to be set for ice server, entry %d", n)
	}

	if len(iceServer.CredentialType) != 0 && iceServer.CredentialType != "password" && iceServer.CredentialType != "oauth" {
		return fmt.Errorf("credential type has to be 'password', 'oauth' ore empty, entry %d", n)
	}
	return nil
}
