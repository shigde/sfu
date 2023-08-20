package lobby

import (
	"testing"

	"github.com/pion/webrtc/v3"
	"github.com/stretchr/testify/assert"
)

func testMessengerSetup(t *testing.T) (*messenger, *senderMock) {
	t.Helper()
	s := newSendMock()
	m := newMessenger(s)
	return m, s
}

func TestMessenger(t *testing.T) {
	t.Run("send Offer", func(t *testing.T) {
		m, sender := testMessengerSetup(t)
		_ = m.sendOffer(*mockedOffer, 2)
		exp := map[int][]uint8{0: {0x7b, 0x22, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x22, 0x3a, 0x32, 0x2c, 0x22, 0x73, 0x64, 0x70, 0x22, 0x3a, 0x7b, 0x22, 0x74, 0x79, 0x70, 0x65, 0x22, 0x3a, 0x22, 0x6f, 0x66, 0x66, 0x65, 0x72, 0x22, 0x2c, 0x22, 0x73, 0x64, 0x70, 0x22, 0x3a, 0x22, 0x2d, 0x2d, 0x6f, 0x2d, 0x2d, 0x22, 0x7d, 0x7d}}
		assert.Equal(t, exp, sender.data)
	})
}

type senderMock struct {
	data     map[int][]byte
	listener func(msg webrtc.DataChannelMessage)
}

func newSendMock() *senderMock {
	return &senderMock{
		data: make(map[int][]byte),
	}
}

func (s *senderMock) OnMessage(f func(msg webrtc.DataChannelMessage)) {
	s.listener = f
}

func (s *senderMock) Send(data []byte) error {
	length := len(s.data)
	s.data[length] = data
	return nil
}
func (s *senderMock) Label() string {
	return "label"
}
