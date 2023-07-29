package lobby

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"golang.org/x/exp/slog"
)

var (
	errRtpSessionAlreadyClosed        = errors.New("the rtp sessions was already closed")
	errReceiverInSessionAlreadyExists = errors.New("receiver already exists")
	errSenderInSessionAlreadyExists   = errors.New("sender already exists")
	errNoSenderInSession              = errors.New("no sender exists")
)

var sessionReqTimeout = 3 * time.Second

type session struct {
	Id               uuid.UUID
	user             uuid.UUID
	rtpEngine        rtpEngine
	hub              *hub
	receiver         receiver
	sender           sender
	sessionReqChan   chan *sessionRequest
	foreignTrackChan chan *hubTrackData
	ownTrackChan     chan *webrtc.TrackLocalStaticRTP
	quit             chan struct{}
}

func newSession(user uuid.UUID, hub *hub, engine rtpEngine) *session {
	quit := make(chan struct{})
	offerChan := make(chan *sessionRequest)
	ownTrackChan := make(chan *webrtc.TrackLocalStaticRTP)

	session := &session{
		Id:               uuid.New(),
		user:             user,
		rtpEngine:        engine,
		hub:              hub,
		sessionReqChan:   offerChan,
		ownTrackChan:     ownTrackChan,
		foreignTrackChan: hub.dispatchChan,
		quit:             quit,
	}

	go session.run()
	return session
}

func (s *session) run() {
	slog.Info("lobby.sessions: run", "id", s.Id, "user", s.user)
	for {
		select {
		case req := <-s.sessionReqChan:
			s.handleSessionReq(req)
		case track := <-s.ownTrackChan:
			s.handleOwnTrack(track)
		case track := <-s.foreignTrackChan:
			s.handleForeignTrack(track)
		case <-s.quit:
			// @TODO Take care that's every stream is closed!
			slog.Info("lobby.sessions: stop running", "id", s.Id, "user", s.user)
			return
		}
	}
}

func (s *session) runRequest(req *sessionRequest) {
	slog.Debug("lobby.sessions: runRequest", "id", s.Id, "user", s.user)
	select {
	case s.sessionReqChan <- req:
		slog.Debug("lobby.sessions: runRequest - return response", "id", s.Id, "user", s.user)
	case <-s.quit:
		req.err <- errRtpSessionAlreadyClosed
		slog.Debug("lobby.sessions: runRequest - interrupted because sessions closed", "id", s.Id, "user", s.user)
	case <-time.After(sessionReqTimeout):
		slog.Error("lobby.sessions: runRequest - interrupted because request timeout", "id", s.Id, "user", s.user)
	}
}

func (s *session) handleSessionReq(req *sessionRequest) {
	slog.Info("lobby.sessions: handle session req", "id", s.Id, "user", s.user)

	var sdp *webrtc.SessionDescription
	var err error
	switch req.sessionReqType {
	case offerReq:
		sdp, err = s.handleOfferReq(req)
	case answerReq:
		sdp, err = s.handleAnswerReq(req)
	case startReq:
		sdp, err = s.handleStartReq(req)
	}
	if err != nil {
		req.err <- fmt.Errorf("handle request: %w", err)
		return
	}

	req.respSDPChan <- sdp
}

func (s *session) handleOfferReq(req *sessionRequest) (*webrtc.SessionDescription, error) {
	if s.receiver != nil {
		return nil, errReceiverInSessionAlreadyExists
	}

	conn, err := s.rtpEngine.NewReceiverConn(*req.reqSDP, s.ownTrackChan)
	if err != nil {
		return nil, fmt.Errorf("create rtp connection: %w", err)
	}
	s.receiver = conn
	answer, err := s.receiver.GetLocalDescription(req.ctx)
	if err != nil {
		return nil, fmt.Errorf("create rtp answerReq: %w", err)
	}
	return answer, nil
}

func (s *session) handleAnswerReq(req *sessionRequest) (*webrtc.SessionDescription, error) {
	if s.sender == nil {
		return nil, errNoSenderInSession
	}
	if err := s.sender.SetAnswer(req.reqSDP); err != nil {
		return nil, fmt.Errorf("setting answer to sender connection: %w", err)
	}
	return nil, nil
}

func (s *session) handleStartReq(req *sessionRequest) (*webrtc.SessionDescription, error) {
	if s.sender != nil {
		return nil, errSenderInSessionAlreadyExists
	}

	var trackList []*webrtc.TrackLocalStaticRTP
	if req.sessionReqType == answerReq {
		trackList = s.hub.getAllTracksFromSessions()
	}

	conn, err := s.rtpEngine.NewSenderConn(trackList)
	if err != nil {
		return nil, fmt.Errorf("create rtp connection: %w", err)
	}

	s.sender = conn

	offer, err := s.sender.GetLocalDescription(req.ctx)
	if err != nil {
		return nil, fmt.Errorf("create rtp answerReq: %w", err)
	}

	return offer, nil
}

func (s *session) handleForeignTrack(track *hubTrackData) {
	if s.sender != nil {
		s.sender.AddTrack(track.track)
	}
}

func (s *session) handleOwnTrack(track *webrtc.TrackLocalStaticRTP) {
	data := &hubTrackData{
		sessionId: s.Id,
		streamId:  track.StreamID(),
		track:     track,
	}
	go func() {
		s.hub.dispatchChan <- data
	}()
}

func (s *session) getTracks() []*webrtc.TrackLocalStaticRTP {
	if s.receiver == nil {
		return nil
	}
	return s.receiver.GetTracks()
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
		<-s.quit
	}
	return nil
}

type receiver interface {
	GetTracks() []*webrtc.TrackLocalStaticRTP
	GetLocalDescription(ctx context.Context) (*webrtc.SessionDescription, error)
}

type sender interface {
	AddTrack(track *webrtc.TrackLocalStaticRTP) bool
	GetLocalDescription(ctx context.Context) (*webrtc.SessionDescription, error)
	SetAnswer(sdp *webrtc.SessionDescription) error
}
