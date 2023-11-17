package cmd

import (
	"fmt"
	"net/url"
	"strings"
)

type shigParams struct {
	Stream   string
	Space    string
	URL      string
	BasePath string
}

func NewShigParamsByUrl(urlString string) (*shigParams, error) {

	shigUrl, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("parsing url: %w", err)
	}
	params := &shigParams{
		URL:      fmt.Sprintf("%s://%s", shigUrl.Scheme, shigUrl.Host),
		BasePath: "",
	}
	prv := ""
	for _, part := range strings.Split(shigUrl.Path, "/") {
		if prv == "space" {
			params.Space = part
		}
		if prv == "stream" {
			params.Stream = part
		}
		prv = strings.ToLower(part)
	}
	return params, nil
}
