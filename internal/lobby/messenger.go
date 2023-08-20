package lobby

import (
	"fmt"
	"sync/atomic"

	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/pkg/message"
	"golang.org/x/exp/slog"
)

type messenger struct {
	counter uint64
	sender  msgSender
}

func newMessenger(s msgSender) *messenger {
	m := &messenger{
		sender: s,
	}
	atomic.AddUint64(&m.counter, 1)
	s.OnMessage(m.onMessages)
	return m
}

func (m *messenger) sendOffer(offer webrtc.SessionDescription, number int) error {
	sdp := &message.Sdp{
		SDP:    offer,
		Number: number,
	}

	msg, err := message.SdpMarshal(sdp)
	if err != nil {
		return fmt.Errorf("marshaling offer message: %w", err)
	}

	if err = m.sender.Send(msg); err != nil {
		return fmt.Errorf("sending offer message: %w", err)
	}

	return nil
}

func (m *messenger) onAnswer(_ webrtc.SessionDescription) error {
	return nil
}

func (m *messenger) onMessages(dcMsg webrtc.DataChannelMessage) {
	if dcMsg.IsString {
		slog.Debug("lobby.messenger: message (string)", "dataChannel", m.sender.Label(), "msg", string(dcMsg.Data))
	} else {
		slog.Debug("lobby.messenger: message ([]byte)", "dataChannel", m.sender.Label(), "length", len(dcMsg.Data))
	}
}

type msgSender interface {
	OnMessage(f func(msg webrtc.DataChannelMessage))
	Send(data []byte) error
	Label() string
}
