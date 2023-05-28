package engine

import "fmt"

type Publisher struct {
	UserId string
	conn   *Connection
}

func NewPublisher(userId string) (*Publisher, error) {
	conn, err := newConnection()
	if err != nil {
		return nil, fmt.Errorf("creation webrtc connection: %w", err)
	}
	return &Publisher{
		UserId: userId,
		conn:   conn,
	}, nil
}

func (peer *Publisher) publish() {

}

func (peer *Publisher) subscribe(_ string) {

}
