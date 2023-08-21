package lobby

import (
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/stretchr/testify/assert"
)

var rawOffer = []byte{0x7b, 0x22, 0x69, 0x64, 0x22, 0x3a, 0x30, 0x2c, 0x22, 0x64, 0x61, 0x74, 0x61, 0x22, 0x3a, 0x7b, 0x22, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x22, 0x3a, 0x32, 0x2c, 0x22, 0x73, 0x64, 0x70, 0x22, 0x3a, 0x7b, 0x22, 0x74, 0x79, 0x70, 0x65, 0x22, 0x3a, 0x22, 0x6f, 0x66, 0x66, 0x65, 0x72, 0x22, 0x2c, 0x22, 0x73, 0x64, 0x70, 0x22, 0x3a, 0x22, 0x2d, 0x2d, 0x6f, 0x2d, 0x2d, 0x22, 0x7d, 0x7d, 0x2c, 0x22, 0x74, 0x79, 0x70, 0x65, 0x22, 0x3a, 0x31, 0x7d}
var rawAnswer = []byte{0x7b, 0x22, 0x69, 0x64, 0x22, 0x3a, 0x32, 0x2c, 0x22, 0x64, 0x61, 0x74, 0x61, 0x22, 0x3a, 0x7b, 0x22, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x22, 0x3a, 0x33, 0x2c, 0x22, 0x73, 0x64, 0x70, 0x22, 0x3a, 0x7b, 0x22, 0x74, 0x79, 0x70, 0x65, 0x22, 0x3a, 0x22, 0x61, 0x6e, 0x73, 0x77, 0x65, 0x72, 0x22, 0x2c, 0x22, 0x73, 0x64, 0x70, 0x22, 0x3a, 0x22, 0x2d, 0x2d, 0x61, 0x2d, 0x2d, 0x22, 0x7d, 0x7d, 0x2c, 0x22, 0x74, 0x79, 0x70, 0x65, 0x22, 0x3a, 0x32, 0x7d}

func testMessengerSetup(t *testing.T) (*messenger, *senderMock, *msgObserverMock) {
	t.Helper()
	s := newSendMock(t)
	m := newMessenger(s)
	o := newMsgObserverMock(t)
	m.register(o)
	s.start()
	return m, s, o
}

func TestMessenger(t *testing.T) {
	t.Run("send Offer", func(t *testing.T) {
		m, sender, _ := testMessengerSetup(t)
		_, _ = m.sendOffer(mockedOffer, 2)
		assert.Equal(t, rawOffer, <-sender.testSendData)
	})

	t.Run("receive Answer", func(t *testing.T) {
		_, sender, o := testMessengerSetup(t)

		var answer *webrtc.SessionDescription
		var index uint32
		var wg sync.WaitGroup
		wg.Add(1)
		o.onAnswerCallback = func(sdp *webrtc.SessionDescription, number uint32) {
			defer wg.Done()
			answer = sdp
			index = number
		}

		sender.updateOnmessageListener(webrtc.DataChannelMessage{Data: rawAnswer})
		wg.Wait()

		assert.Equal(t, mockedAnswer, answer)
		assert.Equal(t, uint32(3), index)
	})
}

type senderMock struct {
	testSendData            chan []byte
	updateOnmessageListener func(msg webrtc.DataChannelMessage)
	start                   func()
}

func (s *senderMock) OnOpen(f func()) {
	s.start = f
}

func newSendMock(t *testing.T) *senderMock {
	t.Helper()
	return &senderMock{
		testSendData: make(chan []byte, 1),
	}
}

func (s *senderMock) OnMessage(f func(msg webrtc.DataChannelMessage)) {
	s.updateOnmessageListener = f
}

func (s *senderMock) Send(data []byte) error {
	s.testSendData <- data
	return nil
}
func (s *senderMock) Label() string {
	return "label"
}

type msgObserverMock struct {
	id               uuid.UUID
	onAnswerCallback func(sdp *webrtc.SessionDescription, number uint32)
}

func newMsgObserverMock(t *testing.T) *msgObserverMock {
	t.Helper()
	return &msgObserverMock{
		id: uuid.New(),
	}
}

func (o *msgObserverMock) onAnswer(sdp *webrtc.SessionDescription, number uint32) {
	if o.onAnswerCallback != nil {
		o.onAnswerCallback(sdp, number)
	}
}

func (o *msgObserverMock) getId() uuid.UUID {
	return o.id
}

func newMockedMessenger(t *testing.T) *messenger {
	t.Helper()
	s := newSendMock(t)
	m := newMessenger(s)
	s.start()
	return m
}
