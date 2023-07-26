package lobby

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/rtp"
	"golang.org/x/exp/slog"
)

var errRtpSessionAlreadyClosed = errors.New("the rtp sessions was already closed")
var errOfferInterrupted = errors.New("request an offer get interrupted")

var sessionReqTimeout = 3 * time.Second

type session struct {
	Id            uuid.UUID
	user          uuid.UUID
	rtpEngine     rtpEngine
	connReceive   *rtp.Connection
	connSend      *rtp.Connection
	offerChan     chan *offerRequest
	onRemoteTrack chan *webrtc.TrackLocalStaticRTP
	onLocalTrack  chan<- *webrtc.TrackLocalStaticRTP
	quit          chan struct{}
}

func newSession(user uuid.UUID, onLocalTrack chan<- *webrtc.TrackLocalStaticRTP, engine rtpEngine) *session {
	quit := make(chan struct{})
	offerChan := make(chan *offerRequest)
	onRemoteTrack := make(chan *webrtc.TrackLocalStaticRTP)
	session := &session{
		Id:            uuid.New(),
		user:          user,
		rtpEngine:     engine,
		offerChan:     offerChan,
		onRemoteTrack: onRemoteTrack,
		onLocalTrack:  onLocalTrack,
		quit:          quit,
	}

	go session.run()
	return session
}

func (s *session) run() {
	slog.Info("lobby.sessions: run", "id", s.Id, "user", s.user)
	for {
		select {
		case offer := <-s.offerChan:
			s.handleOffer(offer)
		case track := <-s.onRemoteTrack:
			s.handleRemoteTrack(track)
		case <-s.quit:
			// @TODO Take care that's every stream is closed!
			slog.Info("lobby.sessions: stop running", "id", s.Id, "user", s.user)
			return
		}
	}
}

func (s *session) runOfferRequest(offerReq *offerRequest) {
	slog.Debug("lobby.sessions: offer", "id", s.Id, "user", s.user)
	select {
	case s.offerChan <- offerReq:
		slog.Debug("lobby.sessions: offer - offer requested", "id", s.Id, "user", s.user)
	case <-s.quit:
		offerReq.err <- errRtpSessionAlreadyClosed
		slog.Debug("lobby.sessions: offer - interrupted because sessions closed", "id", s.Id, "user", s.user)
	case <-time.After(sessionReqTimeout):
		slog.Error("lobby.sessions: offer - interrupted because request timeout", "id", s.Id, "user", s.user)
	}
}

func (s *session) handleOffer(offerReq *offerRequest) {
	slog.Info("lobby.sessions: handle offer", "id", s.Id, "user", s.user)
	conn, err := s.rtpEngine.NewConnection(*offerReq.offer, s.onLocalTrack)

	if err != nil {
		offerReq.err <- fmt.Errorf("create rtp connection: %w", err)
		return
	}

	answer, err := conn.GetAnswer(offerReq.ctx)
	if err != nil {
		offerReq.err <- fmt.Errorf("create rtp answer: %w", err)
		return
	}
	s.connReceive = conn
	offerReq.answer <- answer
}

func (s *session) handleRemoteTrack(track *webrtc.TrackLocalStaticRTP) {
	if s.connSend != nil {
		//s.connSend.sender.addTrack(track)
	}
}

func (s *session) stop() error {
	slog.Info("lobby.sessions: stop", "id", s.Id, "user", s.user)
	select {
	case <-s.quit:
		slog.Error("lobby.sessions: the rtp sessions was already closed", "id", s.Id, "user", s.user)
		return errRtpSessionAlreadyClosed
	default:
		close(s.quit)
		slog.Info("lobby.sessions: stopped was triggered", "id", s.Id, "user", s.user)
	}
	return nil
}
