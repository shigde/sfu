package federation

import (
	"context"

	"github.com/google/uuid"
)

type HostController struct {
	ctx     context.Context
	lobbyId uuid.UUID
	//lobby    *lobby2.lobby
	//hostApi  *clients.ApiClient
	settings *hostSettings
}

//
//func NewHostController(ctx context.Context, lobbyId uuid.UUID, settings *hostSettings) *HostController {
//	hostApi := clients.NewApiClient(settings.instanceId, settings.token, settings.name, settings.url)
//	controller := &HostController{
//		ctx:      ctx,
//		lobbyId:  lobbyId,
//		hostApi:  hostApi,
//		settings: settings,
//	}
//
//	go controller.run()
//	if !settings.isHost {
//		go func() {
//			slog.Debug("lobby.HostController. connect to live stream host instance", "instanceId", settings.instanceId)
//			if _, err := hostApi.Login(); err != nil {
//				slog.Error("login to remote host", "err", err)
//				return
//			}
//
//			if err := controller.connectToHostPipe(settings.instanceId); err != nil {
//				slog.Error("lobby.HostController: connecting pipe", "err", err)
//				return
//			}
//			if err := controller.connectToHostEgress(settings.instanceId); err != nil {
//				slog.Error("lobby.HostController: connecting ingress", "err", err)
//				return
//			}
//		}()
//	}
//	return controller
//}
//
//func (c *HostController) run() {
//	slog.Info("lobby.HostController: run", "lobbyId", c.lobbyId)
//	for {
//		select {
//		case <-c.ctx.Done():
//			return
//		default:
//		}
//	}
//}
//
//func (c *HostController) connectToHostPipe(instanceId uuid.UUID) error {
//	ctx, cancelReq := context.WithCancel(context.Background())
//	defer cancelReq()
//
//	request := lobby2.newLobbyRequest(ctx, instanceId)
//	data := lobby2.newHostGetPipeOfferData()
//	request.data = data
//	go c.lobby.runRequest(request)
//
//	var answer *webrtc.SessionDescription
//	select {
//	case <-c.ctx.Done():
//		return lobby2.errSessionAlreadyClosed
//	case err := <-request.err:
//		return fmt.Errorf("requesting pipe offer: %w", err)
//	case res := <-data.response:
//		var err error
//		if answer, err = c.hostApi.PostHostPipeOffer(c.settings.space, c.settings.stream, res.offer); err != nil {
//			return fmt.Errorf("remote host answer request: %w", err)
//		}
//	}
//	return c.onHostPipeAnswerResponse(answer, instanceId)
//}
//
//func (c *HostController) onHostPipeAnswerResponse(answer *webrtc.SessionDescription, instanceId uuid.UUID) error {
//	ctx, cancelReq := context.WithCancel(context.Background())
//	defer cancelReq()
//	request := lobby2.newLobbyRequest(ctx, instanceId)
//	data := lobby2.newHostSetPipeAnswerData(answer)
//	request.data = data
//	go c.lobby.runRequest(request)
//
//	select {
//	case <-c.ctx.Done():
//		return lobby2.errSessionAlreadyClosed
//	case err := <-request.err:
//		return fmt.Errorf("setting answer for pip connection: %w", err)
//	case connected := <-data.response:
//		slog.Info("lobby.HostController received answer response", "isSuccess", connected)
//		if !connected {
//			return errors.New("pipe connection could not be established")
//		}
//	}
//	return nil
//}
//
//func (c *HostController) connectToHostEgress(instanceId uuid.UUID) error {
//	ctx, cancelReq := context.WithCancel(context.Background())
//	defer cancelReq()
//
//	request := lobby2.newLobbyRequest(ctx, instanceId)
//	data := lobby2.newHostGetEgressOfferData()
//	request.data = data
//	go c.lobby.runRequest(request)
//
//	var answer *webrtc.SessionDescription
//	select {
//	case <-c.ctx.Done():
//		return lobby2.errSessionAlreadyClosed
//	case err := <-request.err:
//		return fmt.Errorf("requesting egress offer: %w", err)
//	case res := <-data.response:
//		var err error
//		if answer, err = c.hostApi.PostHostIngressOffer(c.settings.space, c.settings.stream, res.offer); err != nil {
//			return fmt.Errorf("remote host egress answer request: %w", err)
//		}
//	}
//	return c.onHostEgressAnswerResponse(answer, instanceId)
//}
//
//func (c *HostController) onHostEgressAnswerResponse(answer *webrtc.SessionDescription, instanceId uuid.UUID) error {
//	ctx, cancelReq := context.WithCancel(context.Background())
//	defer cancelReq()
//	request := lobby2.newLobbyRequest(ctx, instanceId)
//	data := lobby2.newHostSetEgressAnswerData(answer)
//	request.data = data
//	go c.lobby.runRequest(request)
//
//	select {
//	case <-c.ctx.Done():
//		return lobby2.errSessionAlreadyClosed
//	case err := <-request.err:
//		return fmt.Errorf("setting answer for pip connection: %w", err)
//	case connected := <-data.response:
//		slog.Info("lobby.HostController received answer response", "isSuccess", connected)
//		if !connected {
//			return errors.New("pipe connection could not be established")
//		}
//	}
//	return nil
//}
//
//func (c *HostController) onRemoteHostPipeConnectionRequest(offer *webrtc.SessionDescription, instanceId uuid.UUID) (*webrtc.SessionDescription, error) {
//	ctx, cancelReq := context.WithCancel(context.Background())
//	defer cancelReq()
//
//	request := lobby2.newLobbyRequest(ctx, instanceId)
//	data := lobby2.newHostGetPipeAnswerData(offer)
//	request.data = data
//	go c.lobby.runRequest(request)
//
//	select {
//	case <-c.ctx.Done():
//		return nil, errors.New("cancel request while the main context is finished")
//	case err := <-request.err:
//		return nil, fmt.Errorf("getting host answer: %w", err)
//	case res := <-data.response:
//		slog.Info("lobby.HostController: getting host answer response success")
//		return res.answer, nil
//	}
//}
//
//func (c *HostController) onRemoteHostIngressConnectionRequest(offer *webrtc.SessionDescription, instanceId uuid.UUID) (*webrtc.SessionDescription, error) {
//	ctx, cancelReq := context.WithCancel(context.Background())
//	defer cancelReq()
//	request := lobby2.newLobbyRequest(ctx, instanceId)
//	data := lobby2.newHostGetIngressAnswerData(offer)
//	request.data = data
//	go c.lobby.runRequest(request)
//
//	select {
//	case <-c.ctx.Done():
//		return nil, errors.New("cancel request while the main context is finished")
//	case err := <-request.err:
//		return nil, fmt.Errorf("getting host ingress answer: %w", err)
//	case res := <-data.response:
//		slog.Info("lobby.HostController: getting host ingress answer response success")
//
//		if c.settings.isHost {
//			go func() {
//				remoteHost, _ := url.Parse("http://stream.localhost:8080")
//				c.hostApi.url = remoteHost
//				slog.Debug("lobby.HostController: start egress endpoint", "host", c.hostApi.url)
//				if _, err := c.hostApi.Login(); err != nil {
//					slog.Error("login to remote host", "err", err)
//					return
//				}
//				if err := c.connectToHostEgress(instanceId); err != nil {
//					slog.Error("lobby.HostController: connecting egress", "err", err)
//					return
//				}
//			}()
//		}
//
//		return res.answer, nil
//	}
//}
