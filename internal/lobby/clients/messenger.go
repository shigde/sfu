package clients

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

type Messenger struct {
	locker       sync.RWMutex
	counter      atomic.Uint32
	sender       msgSender
	observerList map[uuid.UUID]msgObserver
	queueChan    chan []byte
	quit         chan struct{}
}

func NewMessenger(s msgSender) *Messenger {
	m := &Messenger{
		locker:       sync.RWMutex{},
		sender:       s,
		observerList: make(map[uuid.UUID]msgObserver),
		queueChan:    make(chan []byte),
		quit:         make(chan struct{}),
	}
	m.counter.Store(0)
	s.OnMessage(m.onMessages)
	s.OnOpen(func() {
		slog.Debug("lobby.Messenger: sender is open start sending worker")
		go func() {
			for {
				slog.Debug("lobby.Messenger: sending worker running")
				select {
				case byteMsg := <-m.queueChan:
					if err := s.Send(byteMsg); err != nil {
						slog.Error("lobby.Messenger: send message", "err", err)
					}
				case <-m.quit:
					slog.Error("lobby.Messenger: closed")
					return
				}
			}
		}()
	})
	return m
}

func (m *Messenger) SendOffer(offer *webrtc.SessionDescription, number uint32) (uint32, error) {
	slog.Debug("lobby.Messenger: start to send offer", "number", number)
	id := m.count()
	return m.sendSDP(offer, id, number)
}
func (m *Messenger) SendAnswer(sdp *webrtc.SessionDescription, id uint32, number uint32) (uint32, error) {
	slog.Debug("lobby.Messenger: start to send answer", "number", number)
	return m.sendSDP(sdp, id, number)
}

func (m *Messenger) sendSDP(sdp *webrtc.SessionDescription, id uint32, number uint32) (uint32, error) {
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
		return id, fmt.Errorf("marshaling offer message (msgId %d sdp %d): %w", id, number, err)
	}

	select {
	case <-m.quit:
	default:
		select {
		case m.queueChan <- byteMsg:
			slog.Debug("lobby.Messenger: sdp is send", "number", number)
		case <-m.quit:
		}
	}

	return id, nil
}

func (m *Messenger) SendMute(mute *message.Mute) error {
	channelMsg := &message.ChannelMsg{
		Id:   0,
		Type: message.MuteMsg,
		Data: mute,
	}

	byteMsg, err := message.Marshal(channelMsg)
	if err != nil {
		return fmt.Errorf("marshaling mute message: %w", err)
	}

	select {
	case <-m.quit:
	default:
		select {
		case m.queueChan <- byteMsg:
			slog.Debug("lobby.Messenger: mute is send")
		case <-m.quit:
		}
	}

	return nil
}

func (m *Messenger) onMessages(dcMsg webrtc.DataChannelMessage) {
	if dcMsg.IsString {
		slog.Debug("lobby.Messenger: message (string)", "dataChannel", m.sender.Label(), "msg", string(dcMsg.Data))
	} else {
		slog.Debug("lobby.Messenger: message ([]byte)", "dataChannel", m.sender.Label(), "length", len(dcMsg.Data))
		channelMsg, err := message.Unmarshal(dcMsg.Data)
		if err != nil {
			slog.Error("lobby.Messenger: unmarshal message ([]byte)", "dataChannel", m.sender.Label(), "length", len(dcMsg.Data))
		}
		m.notifyAll(channelMsg)
	}
}

func (m *Messenger) Register(o msgObserver) {
	m.locker.Lock()
	defer m.locker.Unlock()
	if _, ok := m.observerList[o.GetId()]; !ok {
		m.observerList[o.GetId()] = o
	}
}

func (m *Messenger) deregister(o msgObserver) {
	m.locker.Lock()
	defer m.locker.Unlock()
	if _, ok := m.observerList[o.GetId()]; ok {
		delete(m.observerList, o.GetId())
	}
}

func (m *Messenger) notifyAll(msg *message.ChannelMsg) {
	switch msg.Type {
	case message.AnswerMsg:
		m.handleAnswerMsg(msg)
	case message.OfferMsg:
		m.handleOfferMsg(msg)
	case message.MuteMsg:
		m.handleMuteMsg(msg)
	default:
		slog.Error("lobby.Messenger: unknown msg type", "err", fmt.Sprintf("unknown msg type: %d", msg.Type), "dataChannel", m.sender.Label())
	}
}

func (m *Messenger) handleAnswerMsg(msg *message.ChannelMsg) {
	answer, err := m.unmarshalSdp(msg)
	if err != nil {
		slog.Error("Messenger: handleAnswerMsg", "err", err)
	}
	m.locker.RLock()
	defer m.locker.RUnlock()
	for _, observer := range m.observerList {
		observer.OnAnswer(answer.SDP, answer.Number)
	}
}

func (m *Messenger) handleOfferMsg(msg *message.ChannelMsg) {
	offer, err := m.unmarshalSdp(msg)
	if err != nil {
		slog.Error("Messenger: handleOfferMsg", "err", err)
	}
	m.locker.RLock()
	defer m.locker.RUnlock()
	for _, observer := range m.observerList {
		observer.OnOffer(offer.SDP, msg.Id, offer.Number)
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

func (m *Messenger) handleMuteMsg(msg *message.ChannelMsg) {
	jsonStr, err := json.Marshal(msg.Data)
	if err != nil {
		slog.Error("lobby.Messenger: unmarshal mute", "err", err, "dataChannel", m.sender.Label())
	}
	mute, err := message.MuteUnmarshal(jsonStr)
	if err != nil {
		slog.Error("lobby.Messenger: unmarshal mute", "err", err, "dataChannel", m.sender.Label())
	}
	slog.Debug("lobby.Messenger: handle incoming mute Msg")

	m.locker.RLock()
	defer m.locker.RUnlock()
	for _, observer := range m.observerList {
		observer.OnMute(mute)
	}
}

func (m *Messenger) close() {
	select {
	case <-m.quit:
	default:
		close(m.quit)
		<-m.quit
	}
}
func (m *Messenger) count() uint32 {
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
	OnAnswer(sdp *webrtc.SessionDescription, number uint32)
	OnOffer(sdp *webrtc.SessionDescription, responseId uint32, responseMsgNumber uint32)
	OnMute(mute *message.Mute)
	GetId() uuid.UUID
}
