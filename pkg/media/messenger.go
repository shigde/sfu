package media

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/pkg/message"
	"golang.org/x/exp/slog"
)

type Messenger struct {
	locker       sync.RWMutex
	observerList map[uuid.UUID]msgObserver
	QueueChan    chan []byte
	quit         chan struct{}
}

func NewMessenger() *Messenger {
	m := &Messenger{
		locker:       sync.RWMutex{},
		observerList: make(map[uuid.UUID]msgObserver),
		QueueChan:    make(chan []byte),
		quit:         make(chan struct{}),
	}
	return m
}

func (m *Messenger) OnMessages(dcMsg webrtc.DataChannelMessage) {
	if !dcMsg.IsString {
		channelMsg, err := message.Unmarshal(dcMsg.Data)
		if err != nil {
			slog.Error("messenger: unmarshal message ([]byte)", "length", len(dcMsg.Data))
		}
		m.notifyAll(channelMsg)
	}
}

func (m *Messenger) notifyAll(msg *message.ChannelMsg) {
	switch msg.Type {
	case message.AnswerMsg:
		m.handleAnswerMsg(msg)
	case message.OfferMsg:
		m.handleOfferMsg(msg)
	default:
		slog.Error("messenger: unknown msg type", "err", fmt.Sprintf("unknown msg type: %d", msg.Type))
	}
}

func (m *Messenger) unmarshalSdp(msg *message.ChannelMsg) (*message.Sdp, error) {
	jsonStr, err := json.Marshal(msg.Data)
	if err != nil {
		return nil, fmt.Errorf("unmarshal message: %w", err)
	}
	sdp, err := message.SdpUnmarshal(jsonStr)
	if err != nil {
		return nil, fmt.Errorf("unmarshal sdp of mesge: %w", err)
	}
	return sdp, nil
}

func (m *Messenger) handleAnswerMsg(msg *message.ChannelMsg) {
	answer, err := m.unmarshalSdp(msg)
	if err != nil {
		slog.Error("messenger: handleAnswerMsg", "err", err)
	}
	m.locker.RLock()
	defer m.locker.RUnlock()
	for _, observer := range m.observerList {
		observer.OnAnswer(answer.SDP, msg.Id, answer.Number)
	}
}

func (m *Messenger) handleOfferMsg(msg *message.ChannelMsg) {
	offer, err := m.unmarshalSdp(msg)
	if err != nil {
		slog.Error("messenger: handleOfferMsg", "err", err)
	}
	m.locker.RLock()
	defer m.locker.RUnlock()
	for _, observer := range m.observerList {
		observer.OnOffer(offer.SDP, msg.Id, offer.Number)
	}
}

func (m *Messenger) SendSDP(sdp *webrtc.SessionDescription, id uint32, number uint32) (uint32, error) {
	sdpMsg := &message.Sdp{
		SDP:    sdp,
		Number: number,
	}

	var msgTye message.MsgType
	switch sdp.Type {
	case webrtc.SDPTypeOffer:
		msgTye = message.OfferMsg
	case webrtc.SDPTypeAnswer:
		msgTye = message.AnswerMsg
	}

	channelMsg := &message.ChannelMsg{
		Id:   id,
		Type: msgTye,
		Data: sdpMsg,
	}

	byteMsg, err := message.Marshal(channelMsg)
	if err != nil {
		return id, fmt.Errorf("marshaling offer message (msgId %d offer %d): %w", id, number, err)
	}

	select {
	case <-m.quit:
	default:
		select {
		case m.QueueChan <- byteMsg:
			slog.Debug("lobby.messenger: offer is send", "number", number)
		case <-m.quit:
		}
	}

	return id, nil
}

func (m *Messenger) Register(o msgObserver) {
	m.locker.Lock()
	defer m.locker.Unlock()
	if _, ok := m.observerList[o.GetId()]; !ok {
		m.observerList[o.GetId()] = o
	}
}

func (m *Messenger) Deregister(o msgObserver) {
	m.locker.Lock()
	defer m.locker.Unlock()
	if _, ok := m.observerList[o.GetId()]; ok {
		delete(m.observerList, o.GetId())
	}
}
