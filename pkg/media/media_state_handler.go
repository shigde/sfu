package media

import (
	"github.com/pion/webrtc/v3"
	"golang.org/x/exp/slog"
)

type MediaStateHandler struct {
	messenger *Messenger
	quit      chan struct{}
}

func NewMediaStateEventHandler(ms *Messenger) *MediaStateHandler {
	return &MediaStateHandler{messenger: ms, quit: make(chan struct{})}
}

func (h *MediaStateHandler) OnConnectionStateChange(state webrtc.ICEConnectionState) {
}
func (h *MediaStateHandler) OnNegotiationNeeded(offer webrtc.SessionDescription) {}
func (h *MediaStateHandler) OnChannel(dc *webrtc.DataChannel) {
	dc.OnOpen(func() {
		dc.OnMessage(h.messenger.OnMessages)
		slog.Debug("messenger: sender is open")
		go func() {
			for {
				slog.Debug("lobby.messenger: sending worker running")
				select {
				case byteMsg := <-h.messenger.QueueChan:
					if err := dc.Send(byteMsg); err != nil {
						slog.Error("lobby.messenger: send message", "err", err)
					}
				case <-h.quit:
					slog.Error("lobby.messenger: closed")
					return
				}
			}
		}()
	})
}

func (h *MediaStateHandler) Close() {
	select {
	case <-h.quit:
	default:
		close(h.quit)
		<-h.quit
	}
}
