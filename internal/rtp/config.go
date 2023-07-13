package rtp

import (
	"fmt"

	"github.com/pion/webrtc/v3"
)

type RtpConfig struct {
	ICEServer []ICEServer `mapstructure:"iceServer"`
}

type ICEServer struct {
	Urls           []string `mapstructure:"urls"`
	Username       string   `mapstructure:"username"`
	Credential     string   `mapstructure:"credential"`
	CredentialType string   `mapstructure:"credentialType"`
}

func (c *RtpConfig) getIceServer() []webrtc.ICEServer {
	iceServerList := []webrtc.ICEServer{}
	for _, server := range c.ICEServer {
		iceServer := webrtc.ICEServer{}
		iceServer.URLs = server.Urls
		iceServer.CredentialType = newICECredentialType(server.CredentialType)
		iceServer.Username = server.Username
		iceServer.Credential = server.Credential
		iceServerList = append(iceServerList, iceServer)
	}

	return iceServerList
}

func (c *RtpConfig) getWebrtcConf() webrtc.Configuration {
	conf := webrtc.Configuration{}
	conf.ICEServers = c.getIceServer()
	return conf
}

func ValidateRtpConfig(config *RtpConfig) error {
	if len(config.ICEServer) != 0 {
		for n, server := range config.ICEServer {
			if err := validateConfigIceServer(server, n); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateConfigIceServer(iceServer ICEServer, n int) error {
	if iceServer.Urls == nil || len(iceServer.Urls) == 0 {
		return fmt.Errorf("rtp.iceServer[]{urls=} has to be set for ice server, entry %d", n)
	}

	if len(iceServer.CredentialType) != 0 && iceServer.CredentialType != "password" && iceServer.CredentialType != "oauth" {
		return fmt.Errorf("rtp.iceServer[]{credentialType=} has to be 'password', 'oauth' ore empty, entry %d", n)
	}
	return nil
}

func newICECredentialType(raw string) webrtc.ICECredentialType {
	switch raw {
	case "oauth":
		return webrtc.ICECredentialTypePassword
	case "password":
		return webrtc.ICECredentialTypeOauth
	default:
		return webrtc.ICECredentialType(webrtc.Unknown)
	}
}
