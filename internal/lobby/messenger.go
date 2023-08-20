package lobby

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/pkg/message"
	"golang.org/x/exp/slog"
)

type messenger struct {
	lock         sync.RWMutex
	counter      uint64
	sender       msgSender
	observerList map[uuid.UUID]msgObserver
}

func newMessenger(s msgSender) *messenger {
	m := &messenger{
		lock:         sync.RWMutex{},
		sender:       s,
		observerList: make(map[uuid.UUID]msgObserver),
	}
	atomic.AddUint64(&m.counter, 1)
	s.OnMessage(m.onMessages)
	return m
}

func (m *messenger) sendOffer(offer webrtc.SessionDescription, number uint64) (uint64, error) {
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

	if err = m.sender.Send(byteMsg); err != nil {
		return id, fmt.Errorf("sending offer message (msgId %d, offer %d) : %w", id, number, err)
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
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.observerList[o.getId()]; !ok {
		m.observerList[o.getId()] = o
	}
}

func (m *messenger) deregister(o msgObserver) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.observerList[o.getId()]; ok {
		delete(m.observerList, o.getId())
	}
}

func (m *messenger) notifyAll(msg *message.ChannelMsg) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	for _, observer := range m.observerList {
		observer.update(*msg)
	}
}

func (m *messenger) count() uint64 {
	return atomic.AddUint64(&m.counter, 1)
}

// -------------- Interfaces ---------- //
type msgSender interface {
	OnMessage(f func(msg webrtc.DataChannelMessage))
	Send(data []byte) error
	Label() string
}

type msgObserver interface {
	update(msg message.ChannelMsg)
	getId() uuid.UUID
}
