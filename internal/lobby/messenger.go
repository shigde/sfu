package lobby

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/pkg/message"
	"golang.org/x/exp/slog"
)

type messenger struct {
	locker       sync.RWMutex
	counter      atomic.Uint32
	sender       msgSender
	observerList map[uuid.UUID]msgObserver
	queueChan    chan []byte
	quit         chan struct{}
}

func newMessenger(s msgSender) *messenger {
	m := &messenger{
		locker:       sync.RWMutex{},
		sender:       s,
		observerList: make(map[uuid.UUID]msgObserver),
		queueChan:    make(chan []byte),
		quit:         make(chan struct{}),
	}
	m.counter.Store(0)
	s.OnMessage(m.onMessages)
	s.OnOpen(func() {
		slog.Debug("lobby.messenger: sender is open start sending worker")
		go func() {
			for {
				slog.Debug("lobby.messenger: sending worker running")
				select {
				case byteMsg := <-m.queueChan:
					if err := s.Send(byteMsg); err != nil {
						slog.Error("lobby.messenger: send message", "err", err)
					}
				case <-m.quit:
					slog.Error("lobby.messenger: closed")
					return
				}
			}
		}()
	})
	return m
}

func (m *messenger) sendOffer(offer *webrtc.SessionDescription, number uint32) (uint32, error) {
	slog.Debug("lobby.messenger: start to send offer", "number", number)
	sdp := &message.Sdp{
		SDP:    offer,
		Number: number,
	}

	id := m.count()
	channelMsg := &message.ChannelMsg{
		Id:   id,
		Type: message.OfferMsg,
		Data: sdp,
	}

	byteMsg, err := message.Marshal(channelMsg)
	if err != nil {
		return id, fmt.Errorf("marshaling offer message (msgId %d offer %d): %w", id, number, err)
	}

	select {
	case <-m.quit:
	default:
		select {
		case m.queueChan <- byteMsg:
			slog.Debug("lobby.messenger: offer is send", "number", number)
		case <-m.quit:
		}
	}

	return id, nil
}

func (m *messenger) onMessages(dcMsg webrtc.DataChannelMessage) {
	if dcMsg.IsString {
		slog.Debug("lobby.messenger: message (string)", "dataChannel", m.sender.Label(), "msg", string(dcMsg.Data))
	} else {
		slog.Debug("lobby.messenger: message ([]byte)", "dataChannel", m.sender.Label(), "length", len(dcMsg.Data))
		channelMsg, err := message.Unmarshal(dcMsg.Data)
		if err != nil {
			slog.Error("lobby.messenger: unmarshal message ([]byte)", "dataChannel", m.sender.Label(), "length", len(dcMsg.Data))
		}
		m.notifyAll(channelMsg)
	}
}

func (m *messenger) register(o msgObserver) {
	m.locker.Lock()
	defer m.locker.Unlock()
	if _, ok := m.observerList[o.getId()]; !ok {
		m.observerList[o.getId()] = o
	}
}

func (m *messenger) deregister(o msgObserver) {
	m.locker.Lock()
	defer m.locker.Unlock()
	if _, ok := m.observerList[o.getId()]; ok {
		delete(m.observerList, o.getId())
	}
}

func (m *messenger) notifyAll(msg *message.ChannelMsg) {
	switch msg.Type {
	case message.AnswerMsg:
		m.handleAnswerMsg(msg)
	default:
		slog.Error("lobby.messenger: unknown msg type", "err", fmt.Sprintf("unknown msg type: %d", msg.Type), "dataChannel", m.sender.Label())
	}
}

func (m *messenger) handleAnswerMsg(msg *message.ChannelMsg) {
	jsonStr, err := json.Marshal(msg.Data)
	if err != nil {
		slog.Error("lobby.messenger: unmarshal answer", "err", err, "dataChannel", m.sender.Label())
	}
	answer, err := message.SdpUnmarshal(jsonStr)
	if err != nil {
		slog.Error("lobby.messenger: unmarshal sdp of answer", "err", err, "dataChannel", m.sender.Label())
	}
	slog.Debug("lobby.messenger: handleAnswerMsg", "number", answer.Number)

	m.locker.RLock()
	defer m.locker.RUnlock()
	for _, observer := range m.observerList {
		observer.onAnswer(answer.SDP, answer.Number)
	}
}

func (m *messenger) close() {
	select {
	case <-m.quit:
	default:
		close(m.quit)
		<-m.quit
	}
}
func (m *messenger) count() uint32 {
	return m.counter.Load()
}

// -------------- Interfaces ---------- //
type msgSender interface {
	OnMessage(f func(msg webrtc.DataChannelMessage))
	Send(data []byte) error
	Label() string
	OnOpen(f func())
}

type msgObserver interface {
	onAnswer(sdp *webrtc.SessionDescription, number uint32)
	getId() uuid.UUID
}
