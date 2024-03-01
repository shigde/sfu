package client

import (
	"net/url"
)

type Client struct {
	Session *Session
	url     *url.URL
}

func NewClient(options ...ClientOption) *Client {
	apiUrl, _ := url.Parse("http://localhost:8080")
	client := &Client{
		Session: &Session{},
		url:     apiUrl,
	}
	for _, opt := range options {
		opt(client)
	}
	return client
}

type ClientOption func(*Client)

func WithSession(session *Session) func(client *Client) {
	return func(client *Client) {
		client.Session = session
	}
}

func WithUrl(apiUrl *url.URL) func(client *Client) {
	return func(client *Client) {
		client.url = apiUrl
	}
}
