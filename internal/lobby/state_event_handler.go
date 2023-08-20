package lobby

import (
	"github.com/pion/webrtc/v3"
)

type stateEventHandler struct {
}

func newStateEventHandler() *stateEventHandler {
	return &stateEventHandler{}
}

func (h *stateEventHandler) OnConnectionStateChange(state webrtc.ICEConnectionState) {}
func (h *stateEventHandler) OnNegotiationNeeded(offer webrtc.SessionDescription)     {}
func (h *stateEventHandler) OnOnChannel(dc *webrtc.DataChannel)                      {}

//func () {
//	// log.Printf("OnOpen: %s-%d. Random messages will now be sent to any connected DataChannels every second\n", dc.Label(), dc.ID())
//
//	for range time.NewTicker(1000 * time.Millisecond).C {
//		// log.Printf("Sending (%d) msg with len %d \n", msgID, len(buf))
//		msgID++
//
//		_ = dc.Send(buf)
//
//	}
//}

//func(i webrtc.ICEConnectionState) {
//	// @TODO Implement irregular connection closed by client handling
//	if i == webrtc.ICEConnectionStateFailed {
//		if err := peerConnection.Close(); err != nil {
//			slog.Error("rtp.engine: receiver peerConnection.Close", "err", err)
//		}
//		receiver.stop()
//	}
//}

// Datachannel messages

//slog.Debug("rtp.engine: data channel '%s'-'%d' open. keep-alive messages will now be sent", d.Label(), d.ID())
//
//for range time.NewTicker(5 * time.Second).C {
//message := "keep-alive"
//sendErr := d.SendText(message)
//if sendErr != nil {
//slog.Warn("rtp.engine: data channel send", "err", sendErr)
//}
//}

//
//d.OnOpen(func() {
//	slog.Debug("rtp.engine: data channel '%s'-'%d' open. keep-alive messages will now be sent", d.Label(), d.ID())
//	handler.OnOnChannel(d)
//	for range time.NewTicker(5 * time.Second).C {
//		message := "keep-alive"
//		sendErr := d.SendText(message)
//		if sendErr != nil {
//			slog.Warn("rtp.engine: data channel send", "err", sendErr)
//		}
//	}
//})
//
//// Register text message handling
//d.OnMessage(func(msg webrtc.DataChannelMessage) {
//	// do something
//	fmt.Printf("Message from DataChannel '%s': '%s'\n", d.Label(), string(msg.Data))
//})
