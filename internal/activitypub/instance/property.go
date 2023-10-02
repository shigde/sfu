package instance

import (
	"fmt"
	"net/url"
)

type Property struct {
	InstanceUrl      *url.URL
	InstanceUsername string
}

func NewProperty(config *FederationConfig) *Property {
	protocol := "http"
	if config.Https {
		protocol = protocol + "s"
	}
	instanceUrl, _ := url.Parse(fmt.Sprintf("%s://%s", protocol, config.Domain))
	return &Property{
		InstanceUrl:      instanceUrl,
		InstanceUsername: config.InstanceUsername,
	}
}
