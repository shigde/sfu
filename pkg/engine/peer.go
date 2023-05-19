package engine

import "fmt"

type Peer struct {
	UserId string
	conn   *Connection
}

func NewPeer(userId string) (*Peer, error) {
	conn, err := newConnection()
	if err != nil {
		return nil, fmt.Errorf("creation webrtc connection: %w", err)
	}
	return &Peer{
		UserId: userId,
		conn:   conn,
	}, nil
}

func (peer *Peer) publish() {

}

func (peer *Peer) subscribe(_ string) {

}
