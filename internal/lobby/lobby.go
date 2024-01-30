package lobby

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/metric"
	"github.com/shigde/sfu/internal/rtmp"
	"github.com/shigde/sfu/internal/rtp"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/slog"
)

var (
	errNoSession            = errors.New("no session exists")
	ErrSessionAlreadyExists = errors.New("session already exists")
	errLobbyStopped         = errors.New("error because lobby stopped")
	errLobbyRequestTimeout  = errors.New("lobby request timeout error")
	lobbyReqTimeout         = 3 * time.Second
)

type lobby struct {
	ctx                   context.Context
	Id                    uuid.UUID
	sessions              *sessionRepository
	liveStreamSender      liveStreamSender
	streamer              *rtmp.Streamer
	hub                   *hub
	isHost                bool
	rtpEngine             rtpEngine
	resourceId            uuid.UUID
	entity                *LobbyEntity
	hostController        *hostInstanceController
	stopRunning           func()
	reqChan               chan *lobbyRequest
	sessionQuit           chan uuid.UUID
	lobbyGarbageCollector chan<- uuid.UUID
}

func newLobby(id uuid.UUID, entity *LobbyEntity, rtpEngine rtpEngine, lobbyGarbageCollector chan<- uuid.UUID, settings hostInstanceSettings) *lobby {
	ctx, stop := context.WithCancel(context.Background())
	sessions := newSessionRepository()
	reqChan := make(chan *lobbyRequest)
	childQuitChan := make(chan uuid.UUID)
	liveStreamSender, err := rtp.NewLiveStreamSender(ctx, id)
	if err != nil {
		slog.Error("create live stream sender", "err", err)
	}

	streamer := rtmp.NewStreamer(ctx)
	hub := newHub(ctx, sessions, entity.LiveStreamId, liveStreamSender)
	lobby := &lobby{
		ctx:                   ctx,
		Id:                    id,
		resourceId:            uuid.New(),
		rtpEngine:             rtpEngine,
		sessions:              sessions,
		liveStreamSender:      liveStreamSender,
		streamer:              streamer,
		hub:                   hub,
		entity:                entity,
		stopRunning:           stop,
		reqChan:               reqChan,
		sessionQuit:           childQuitChan,
		lobbyGarbageCollector: lobbyGarbageCollector,
	}
	go lobby.run()
	lobby.hostController = newHostInstanceController(ctx, id, lobby, settings)
	return lobby
}

func (l *lobby) run() {
	slog.Info("lobby.lobby: run", "lobbyId", l.Id)

	for {
		select {
		case req := <-l.reqChan:
			switch requestType := req.data.(type) {
			// This needs another pattern, much redundant code
			case *createIngressEndpointData:
				l.handleCreateIngressEndpoint(req)
			case *initEgressEndpointData:
				l.handleInitEgressEndpoint(req)
			case *finalCreateEgressEndpointData:
				l.handleFinalCreateEgressEndpointData(req)
			case *createMainEgressEndpointData:
				l.handleCreateMainEgressEndpointData(req)
			case *leaveData:
				l.handleLeave(req)

			case *liveStreamData:
				l.handleLiveStreamReq(req)

			// Server to Server	Data Channel
			case *hostGetPipeOfferData:
				l.handleOfferHostPipeEndpoint(req)
			case *hostSetPipeAnswerData:
				l.handleAnswerHostPipeEndpoint(req)
			case *hostGetPipeAnswerData:
				l.handleHostRemotePipeEndpoint(req)

			// Server to Server Media
			case *hostGetEgressOfferData:
				l.handleOfferHostEgressEndpoint(req)
			case *hostSetEgressAnswerData:
				l.handleAnswerHostEgressEndpoint(req)
			case *hostGetIngressAnswerData:
				l.handleHostIngressEndpoint(req)
			default:
				slog.Error("lobby.lobby: not supported request type in lobby", "type", requestType)
			}
		case id := <-l.sessionQuit:
			slog.Debug("lobby.lobby: session quit lobby", "session", id)
			if _, err := l.deleteSessionByUserId(id); err != nil {
				slog.Error("lobby.lobby: deleting session because internally reason", "err", err, "session", id)
			}
		case <-l.ctx.Done():
			slog.Info("lobby.lobby: close lobby", "lobbyId", l.Id)
			return
		}
	}
}

// @TODO Refactor this for better understanding
// Maybe when I use an error chanel as return and fill the pointer of result with result value I could simplify this
// methode and even close the channels more safety.
// Even the errors of runRequest will not mixed withe the errors resulting by the request command.
// ...
//
//	func (l *lobby) runRequest(req *lobbyRequest) <-error {
//	   err := make(chanel error)
//	   defer close(err)
//	   ...
//	   return err
//	}
//
// Open Question: But before I want do this I have to find a way that's make the calling function waiting for the result
// of the request command.
func (l *lobby) runRequest(req *lobbyRequest) {
	slog.Debug("lobby.lobby: runRequest", "lobbyId", l.Id, "user", req.user)
	select {
	case l.reqChan <- req:
		slog.Debug("lobby.lobby: runRequest - request finish", "lobbyId", l.Id, "user", req.user)
	case <-l.ctx.Done():
		req.err <- errLobbyStopped
		slog.Debug("lobby.lobby: runRequest - interrupted because lobby closed", "lobbyId", l.Id, "user", req.user)
	case <-time.After(lobbyReqTimeout):
		req.err <- errLobbyRequestTimeout
		slog.Error("lobby.lobby: runRequest - interrupted because request timeout", "lobbyId", l.Id, "user", req.user)
	}
}

func (l *lobby) handleCreateIngressEndpoint(lobbyReq *lobbyRequest) {
	slog.Info("lobby.lobby: handle join", "lobbyId", l.Id, "user", lobbyReq.user)
	ctx, span := otel.Tracer(tracerName).Start(lobbyReq.ctx, "lobby:handleCreateIngressEndpoint")
	lobbyReq.ctx = ctx
	defer span.End()

	data, _ := lobbyReq.data.(*createIngressEndpointData)
	session, ok := l.sessions.FindByUserId(lobbyReq.user)
	if ok {
		select {
		case lobbyReq.err <- ErrSessionAlreadyExists:
		case <-lobbyReq.ctx.Done():
			lobbyReq.err <- errLobbyRequestTimeout
		case <-l.ctx.Done():
			lobbyReq.err <- errLobbyStopped
		}
		return
	}
	session = newSession(lobbyReq.user, l.hub, l.rtpEngine, l.sessionQuit)
	l.sessions.Add(session)
	metric.RunningSessionsInc(l.Id.String())
	offerReq := newSessionRequest(lobbyReq.ctx, data.offer, offerIngressReq)

	go func() {
		slog.Info("lobby.lobby: create offerIngressReq request", "lobbyId", l.Id, "user", lobbyReq.user)
		session.runRequest(offerReq)
	}()

	select {
	case answer := <-offerReq.respSDPChan:
		data.response <- &createIngressEndpointResponse{
			answer:       answer,
			resource:     l.resourceId,
			RtpSessionId: session.Id,
		}
	case err := <-offerReq.err:
		lobbyReq.err <- fmt.Errorf("start session for joing: %w", err)
	case <-lobbyReq.ctx.Done():
		lobbyReq.err <- errLobbyRequestTimeout
	case <-l.ctx.Done():
		lobbyReq.err <- errLobbyStopped
	}
}

func (l *lobby) handleInitEgressEndpoint(req *lobbyRequest) {
	slog.Info("lobby.lobby: handle start listen", "lobbyId", l.Id, "user", req.user)
	ctx, span := otel.Tracer(tracerName).Start(req.ctx, "lobby:handleInitEgressEndpoint")
	req.ctx = ctx
	defer span.End()

	data, _ := req.data.(*initEgressEndpointData)

	session, ok := l.sessions.FindByUserId(req.user)
	if !ok {
		req.err <- errNoSession
		return
	}
	startSessionReq := newSessionRequest(ctx, nil, initEgressReq)

	go func() {
		slog.Info("lobby.lobby: create offerIngressReq request", "lobbyId", l.Id, "user", req.user)
		session.runRequest(startSessionReq)
	}()
	select {
	case offer := <-startSessionReq.respSDPChan:
		data.response <- &initEgressEndpointResponse{
			offer:        offer,
			RtpSessionId: session.Id,
		}
	case err := <-startSessionReq.err:
		req.err <- fmt.Errorf("start session for listening: %w", err)
	case <-req.ctx.Done():
		req.err <- errLobbyRequestTimeout
	case <-l.ctx.Done():
		req.err <- errLobbyStopped
	}
}

func (l *lobby) handleFinalCreateEgressEndpointData(req *lobbyRequest) {
	slog.Info("lobby.lobby: handle listen", "lobbyId", l.Id, "user", req.user)
	ctx, span := otel.Tracer(tracerName).Start(req.ctx, "lobby:handleFinalCreateEgressEndpointData")
	req.ctx = ctx
	defer span.End()

	data, _ := req.data.(*finalCreateEgressEndpointData)
	session, ok := l.sessions.FindByUserId(req.user)
	if !ok {
		select {
		case req.err <- fmt.Errorf("no session existing for user %s: %w", req.user, errNoSession):
		case <-req.ctx.Done():
			req.err <- errLobbyRequestTimeout
		case <-l.ctx.Done():
			req.err <- errLobbyStopped
		}
		return
	}

	answerReq := newSessionRequest(req.ctx, data.answer, answerEgressReq)
	go func() {
		slog.Info("lobby.lobby: create offerIngressReq request", "lobbyId", l.Id, "user", req.user)
		session.runRequest(answerReq)
	}()

	select {
	case _ = <-answerReq.respSDPChan:
		data.response <- &finalCreateEgressEndpointResponse{
			RtpSessionId: session.Id,
		}
	case err := <-answerReq.err:
		req.err <- fmt.Errorf("listening on session: %w", err)
	case <-req.ctx.Done():
		req.err <- errLobbyRequestTimeout
	case <-l.ctx.Done():
		req.err <- errLobbyStopped
	}
}

func (l *lobby) handleCreateMainEgressEndpointData(lobbyReq *lobbyRequest) {
	slog.Info("lobby.lobby: handle main egress", "lobbyId", l.Id, "user", lobbyReq.user)
	ctx, span := otel.Tracer(tracerName).Start(lobbyReq.ctx, "lobby:handleCreateMainEgressEndpointData")
	lobbyReq.ctx = ctx
	defer span.End()

	data, _ := lobbyReq.data.(*createMainEgressEndpointData)
	session, ok := l.sessions.FindByUserId(lobbyReq.user)
	if ok {
		select {
		case lobbyReq.err <- ErrSessionAlreadyExists:
		case <-lobbyReq.ctx.Done():
			lobbyReq.err <- errLobbyRequestTimeout
		case <-l.ctx.Done():
			lobbyReq.err <- errLobbyStopped
		}
		return
	}
	session = newSession(lobbyReq.user, l.hub, l.rtpEngine, l.sessionQuit)
	l.sessions.Add(session)
	offerReq := newSessionRequest(lobbyReq.ctx, data.offer, offerIngressReq)

	go func() {
		slog.Info("lobby.lobby: create offerIngressReq request", "lobbyId", l.Id, "user", lobbyReq.user)
		session.runRequest(offerReq)
	}()
	select {
	case answer := <-offerReq.respSDPChan:
		data.response <- &createMainEgressEndpointResponse{
			answer:       answer,
			RtpSessionId: session.Id,
		}
	case err := <-offerReq.err:
		lobbyReq.err <- fmt.Errorf("start session for joing: %w", err)
	case <-lobbyReq.ctx.Done():
		lobbyReq.err <- errLobbyRequestTimeout
	case <-l.ctx.Done():
		lobbyReq.err <- errLobbyStopped
	}
}

func (l *lobby) handleLeave(req *lobbyRequest) {
	slog.Info("lobby.lobby: handleLeave", "lobbyId", l.Id, "user", req.user)
	data, _ := req.data.(*leaveData)

	deleted, err := l.deleteSessionByUserId(req.user)
	if err != nil {
		req.err <- fmt.Errorf("no session existing for user %s: %w", req.user, errNoSession)
	}
	data.response <- deleted
}

func (l *lobby) handleLiveStreamReq(req *lobbyRequest) {
	slog.Info("lobby.lobby: handleLiveStreamReq", "lobbyId", l.Id, "user", req.user)
	data, _ := req.data.(*liveStreamData)
	if data.cmd == "start" {
		slog.Debug("lobby.lobby: start ffmpeg sender", "lobbyId", l.Id, "user", req.user)
		streamUrl := fmt.Sprintf("%s/%s", data.rtmpUrl, data.key)
		if err := l.streamer.StartFFmpeg(streamUrl); err != nil {
			req.err <- fmt.Errorf("starting ffmeg: %w", err)
			return
		}
	}
	if data.cmd == "stop" {
		slog.Debug("lobby.lobby: stop ffmpeg sender", "lobbyId", l.Id, "user", req.user)
		oldStreamer := l.streamer
		l.streamer = rtmp.NewStreamer(l.ctx)
		oldStreamer.Stop()
	}

	//if err != nil {
	//	req.err <- fmt.Errorf("no session existing for user %s: %w", req.user, errNoSession)
	//}
	data.response <- true
}

func (l *lobby) stop() {
	slog.Info("lobby.lobby: stop", "lobbyId", l.Id)
	select {
	case <-l.ctx.Done():
		slog.Warn("lobby.lobby: the lobby was already closed", "lobbyId", l.Id)
	default:
		l.stopRunning()
		slog.Info("lobby.lobby: stopped was triggered", "lobbyId", l.Id)
	}
}

func (l *lobby) deleteSessionByUserId(userId uuid.UUID) (bool, error) {
	if session, ok := l.sessions.FindByUserId(userId); ok {
		deleted := l.sessions.Delete(session.Id)
		slog.Debug("lobby.lobby: deleteSessionByUserId", "lobbyId", l.Id, "sessionId", session.Id, "userId", userId, "deleted", deleted)
		session.stop()
		metric.RunningSessionsDec(l.Id.String())

		// When Lobby is empty then it is time to close the lobby.
		// But we have to take care about races, because in the meanwhile a new session request could be made.
		// We leave to the lobby manager the dealing with races.
		if l.sessions.Len() == 0 {
			// Spawn a routine to not block the lobby process at all
			go func() {
				l.lobbyGarbageCollector <- l.Id
			}()
		}
		return deleted, nil
	}
	return false, errNoSession
}

func (l *lobby) handleOfferHostEgressEndpoint(lobbyReq *lobbyRequest) {
	slog.Info("lobby.lobby: handle start offer host egress egress", "lobbyId", l.Id, "user", lobbyReq.user)
	ctx, span := otel.Tracer(tracerName).Start(lobbyReq.ctx, "lobby:handleOfferHostEgressEndpoint")
	lobbyReq.ctx = ctx
	defer span.End()

	data, _ := lobbyReq.data.(*hostGetEgressOfferData)

	session, ok := l.sessions.FindByUserId(lobbyReq.user)
	if !ok {
		select {
		case lobbyReq.err <- fmt.Errorf("no session existing for instance %s: %w", lobbyReq.user, errNoSession):
		case <-lobbyReq.ctx.Done():
			lobbyReq.err <- errLobbyRequestTimeout
		case <-l.ctx.Done():
			lobbyReq.err <- errLobbyStopped
		}
		return
	}

	startSessionReq := newSessionRequest(ctx, nil, offerHostEgressReq)

	go func() {
		slog.Info("lobby.lobby: create offerHostEgressReq request", "lobbyId", l.Id, "user", lobbyReq.user)
		session.runRequest(startSessionReq)
	}()
	select {
	case offer := <-startSessionReq.respSDPChan:
		data.response <- &hostOfferResponse{
			offer:        offer,
			RtpSessionId: session.Id,
		}
	case err := <-startSessionReq.err:
		lobbyReq.err <- fmt.Errorf("handling offer host egress egress : %w", err)

	case <-lobbyReq.ctx.Done():
		lobbyReq.err <- errLobbyRequestTimeout
	case <-l.ctx.Done():
		lobbyReq.err <- errLobbyStopped
	}
}

func (l *lobby) handleAnswerHostEgressEndpoint(lobbyReq *lobbyRequest) {
	slog.Info("lobby.lobby: handle answer host egress egress", "lobbyId", l.Id, "instance", lobbyReq.user)
	ctx, span := otel.Tracer(tracerName).Start(lobbyReq.ctx, "lobby:handleFinalCreateEgressEndpointData")
	lobbyReq.ctx = ctx
	defer span.End()

	data, _ := lobbyReq.data.(*hostSetEgressAnswerData)
	session, ok := l.sessions.FindByUserId(lobbyReq.user)
	if !ok {
		select {
		case lobbyReq.err <- fmt.Errorf("no session existing for instance %s: %w", lobbyReq.user, errNoSession):
		case <-lobbyReq.ctx.Done():
			lobbyReq.err <- errLobbyRequestTimeout
		case <-l.ctx.Done():
			lobbyReq.err <- errLobbyStopped
		}
		return
	}

	answerReq := newSessionRequest(lobbyReq.ctx, data.answer, answerHostEgressReq)
	go func() {
		slog.Info("lobby.lobby: create answerHostEgressReq request", "lobbyId", l.Id, "instanceId", lobbyReq.user)
		session.runRequest(answerReq)
	}()

	select {
	case _ = <-answerReq.respSDPChan:
		data.response <- true
	case err := <-answerReq.err:
		lobbyReq.err <- fmt.Errorf("handle answer host egress egress: %w", err)
	case <-lobbyReq.ctx.Done():
		lobbyReq.err <- errLobbyRequestTimeout
	case <-l.ctx.Done():
		lobbyReq.err <- errLobbyStopped
	}
}

func (l *lobby) handleHostIngressEndpoint(lobbyReq *lobbyRequest) {
	slog.Info("lobby.lobby: handle host ingress egress", "lobbyId", l.Id, "instanceId", lobbyReq.user)
	ctx, span := otel.Tracer(tracerName).Start(lobbyReq.ctx, "lobby:handleHostIngressEndpoint")
	lobbyReq.ctx = ctx
	defer span.End()

	data, _ := lobbyReq.data.(*hostGetIngressAnswerData)
	session, ok := l.sessions.FindByUserId(lobbyReq.user)
	if !ok {
		session = newSession(lobbyReq.user, l.hub, l.rtpEngine, l.sessionQuit)
		l.sessions.Add(session)
		metric.RunningSessionsInc(l.Id.String())
	}

	offerReq := newSessionRequest(lobbyReq.ctx, data.offer, offerHostIngressReq)

	go func() {
		slog.Info("lobby.lobby: create offerHostIngressReq request", "lobbyId", l.Id, "instanceId", lobbyReq.user)
		session.runRequest(offerReq)
	}()

	select {
	case answer := <-offerReq.respSDPChan:
		data.response <- &hostAnswerResponse{
			answer:       answer,
			RtpSessionId: session.Id,
		}
	case err := <-offerReq.err:
		lobbyReq.err <- fmt.Errorf("handling host ingress egress: %w", err)
	case <-lobbyReq.ctx.Done():
		lobbyReq.err <- errLobbyRequestTimeout
	case <-l.ctx.Done():
		lobbyReq.err <- errLobbyStopped
	}
}

func (l *lobby) handleOfferHostPipeEndpoint(lobbyReq *lobbyRequest) {
	slog.Info("lobby.lobby: handle start offer host pipe egress", "lobbyId", l.Id, "user", lobbyReq.user)
	ctx, span := otel.Tracer(tracerName).Start(lobbyReq.ctx, "lobby:handleOfferHostPipeEndpoint")
	lobbyReq.ctx = ctx
	defer span.End()

	data, _ := lobbyReq.data.(*hostGetPipeOfferData)

	session, ok := l.sessions.FindByUserId(lobbyReq.user)
	if ok {
		lobbyReq.err <- ErrSessionAlreadyExists
		return
	}
	session = newSession(lobbyReq.user, l.hub, l.rtpEngine, l.sessionQuit)
	l.sessions.Add(session)
	startSessionReq := newSessionRequest(ctx, nil, offerHostPipeReq)

	go func() {
		slog.Info("lobby.lobby: create offerHostPipeReq request", "lobbyId", l.Id, "user", lobbyReq.user)
		session.runRequest(startSessionReq)
	}()
	select {
	case offer := <-startSessionReq.respSDPChan:
		data.response <- &hostOfferResponse{
			offer:        offer,
			RtpSessionId: session.Id,
		}
	case err := <-startSessionReq.err:
		lobbyReq.err <- fmt.Errorf("handling offer host egress egress : %w", err)

	case <-lobbyReq.ctx.Done():
		lobbyReq.err <- errLobbyRequestTimeout
	case <-l.ctx.Done():
		lobbyReq.err <- errLobbyStopped
	}
}

func (l *lobby) handleAnswerHostPipeEndpoint(lobbyReq *lobbyRequest) {
	slog.Info("lobby.lobby: handle answer host pipe egress", "lobbyId", l.Id, "instance", lobbyReq.user)
	ctx, span := otel.Tracer(tracerName).Start(lobbyReq.ctx, "lobby:handleAnswerHostPipeEndpoint")
	lobbyReq.ctx = ctx
	defer span.End()

	data, _ := lobbyReq.data.(*hostSetPipeAnswerData)
	session, ok := l.sessions.FindByUserId(lobbyReq.user)
	if !ok {
		select {
		case lobbyReq.err <- fmt.Errorf("no session existing for instance %s: %w", lobbyReq.user, errNoSession):
		case <-lobbyReq.ctx.Done():
			lobbyReq.err <- errLobbyRequestTimeout
		case <-l.ctx.Done():
			lobbyReq.err <- errLobbyStopped
		}
		return
	}

	answerReq := newSessionRequest(lobbyReq.ctx, data.answer, answerHostPipeReq)
	go func() {
		slog.Info("lobby.lobby: create answerHostPipeReq request", "lobbyId", l.Id, "instanceId", lobbyReq.user)
		session.runRequest(answerReq)
	}()

	select {
	case _ = <-answerReq.respSDPChan:
		data.response <- true
	case err := <-answerReq.err:
		lobbyReq.err <- fmt.Errorf("handle answer host egress egress: %w", err)
	case <-lobbyReq.ctx.Done():
		lobbyReq.err <- errLobbyRequestTimeout
	case <-l.ctx.Done():
		lobbyReq.err <- errLobbyStopped
	}
}

func (l *lobby) handleHostRemotePipeEndpoint(lobbyReq *lobbyRequest) {
	slog.Info("lobby.lobby: handle host remote pipe egress", "lobbyId", l.Id, "instanceId", lobbyReq.user)
	ctx, span := otel.Tracer(tracerName).Start(lobbyReq.ctx, "lobby:handleHostRemotePipeEndpoint")
	lobbyReq.ctx = ctx
	defer span.End()

	data, _ := lobbyReq.data.(*hostGetPipeAnswerData)
	session, ok := l.sessions.FindByUserId(lobbyReq.user)
	if !ok {
		session = newSession(lobbyReq.user, l.hub, l.rtpEngine, l.sessionQuit)
		l.sessions.Add(session)
		metric.RunningSessionsInc(l.Id.String())
	}

	offerReq := newSessionRequest(lobbyReq.ctx, data.offer, offerHostRemotePipeReq)

	go func() {
		slog.Info("lobby.lobby: create hostGetPipeAnswerData request", "lobbyId", l.Id, "instanceId", lobbyReq.user)
		session.runRequest(offerReq)
	}()

	select {
	case answer := <-offerReq.respSDPChan:
		data.response <- &hostAnswerResponse{
			answer:       answer,
			RtpSessionId: session.Id,
		}
	case err := <-offerReq.err:
		lobbyReq.err <- fmt.Errorf("handling host ingress egress: %w", err)
	case <-lobbyReq.ctx.Done():
		lobbyReq.err <- errLobbyRequestTimeout
	case <-l.ctx.Done():
		lobbyReq.err <- errLobbyStopped
	}
}
