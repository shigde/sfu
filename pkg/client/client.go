package client

type Client struct {
	Session *Session
}

func NewClient(options ...ClientOption) *Client {
	client := &Client{
		Session: &Session{},
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
