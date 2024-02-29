package lobby

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"golang.org/x/exp/slog"
)

type hostInstanceController struct {
	ctx      context.Context
	lobbyId  uuid.UUID
	lobby    *lobby
	hostApi  *hostInstanceApiClient
	settings hostInstanceSettings
}

func newHostInstanceController(ctx context.Context, lobbyId uuid.UUID, lobby *lobby, settings hostInstanceSettings) *hostInstanceController {
	hostApi := NewHostInstanceApiClient(settings.instanceId, settings.token, settings.name, settings.url)
	controller := &hostInstanceController{
		ctx:      ctx,
		lobbyId:  lobbyId,
		lobby:    lobby,
		hostApi:  hostApi,
		settings: settings,
	}

	go controller.run()
	if !settings.isHost {
		go func() {
			slog.Debug("lobby.hostInstanceController. connect to live stream host instance", "instanceId", settings.instanceId)
			if _, err := hostApi.Login(); err != nil {
				slog.Error("login to remote host", "err", err)
				return
			}

			if err := controller.connectToHostPipe(settings.instanceId); err != nil {
				slog.Error("lobby.hostInstanceController: connecting pipe", "err", err)
				return
			}
			if err := controller.connectToHostEgress(settings.instanceId); err != nil {
				slog.Error("lobby.hostInstanceController: connecting ingress", "err", err)
				return
			}
		}()
	}
	return controller
}

func (c *hostInstanceController) run() {
	slog.Info("lobby.hostInstanceController: run", "lobbyId", c.lobbyId)
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}
	}
}

func (c *hostInstanceController) connectToHostPipe(instanceId uuid.UUID) error {
	ctx, cancelReq := context.WithCancel(context.Background())
	defer cancelReq()

	request := newLobbyRequest(ctx, instanceId)
	data := newHostGetPipeOfferData()
	request.data = data
	go c.lobby.runRequest(request)

	var answer *webrtc.SessionDescription
	select {
	case <-c.ctx.Done():
		return errSessionAlreadyClosed
	case err := <-request.err:
		return fmt.Errorf("requesting pipe offer: %w", err)
	case res := <-data.response:
		var err error
		if answer, err = c.hostApi.PostHostPipeOffer(c.settings.space, c.settings.stream, res.offer); err != nil {
			return fmt.Errorf("remote host answer request: %w", err)
		}
	}
	return c.onHostPipeAnswerResponse(answer, instanceId)
}

func (c *hostInstanceController) onHostPipeAnswerResponse(answer *webrtc.SessionDescription, instanceId uuid.UUID) error {
	ctx, cancelReq := context.WithCancel(context.Background())
	defer cancelReq()
	request := newLobbyRequest(ctx, instanceId)
	data := newHostSetPipeAnswerData(answer)
	request.data = data
	go c.lobby.runRequest(request)

	select {
	case <-c.ctx.Done():
		return errSessionAlreadyClosed
	case err := <-request.err:
		return fmt.Errorf("setting answer for pip connection: %w", err)
	case connected := <-data.response:
		slog.Info("lobby.hostInstanceController received answer response", "isSuccess", connected)
		if !connected {
			return errors.New("pipe connection could not be established")
		}
	}
	return nil
}

func (c *hostInstanceController) connectToHostEgress(instanceId uuid.UUID) error {
	ctx, cancelReq := context.WithCancel(context.Background())
	defer cancelReq()

	request := newLobbyRequest(ctx, instanceId)
	data := newHostGetEgressOfferData()
	request.data = data
	go c.lobby.runRequest(request)

	var answer *webrtc.SessionDescription
	select {
	case <-c.ctx.Done():
		return errSessionAlreadyClosed
	case err := <-request.err:
		return fmt.Errorf("requesting egress offer: %w", err)
	case res := <-data.response:
		var err error
		if answer, err = c.hostApi.PostHostIngressOffer(c.settings.space, c.settings.stream, res.offer); err != nil {
			return fmt.Errorf("remote host egress answer request: %w", err)
		}
	}
	return c.onHostEgressAnswerResponse(answer, instanceId)
}

func (c *hostInstanceController) onHostEgressAnswerResponse(answer *webrtc.SessionDescription, instanceId uuid.UUID) error {
	ctx, cancelReq := context.WithCancel(context.Background())
	defer cancelReq()
	request := newLobbyRequest(ctx, instanceId)
	data := newHostSetEgressAnswerData(answer)
	request.data = data
	go c.lobby.runRequest(request)

	select {
	case <-c.ctx.Done():
		return errSessionAlreadyClosed
	case err := <-request.err:
		return fmt.Errorf("setting answer for pip connection: %w", err)
	case connected := <-data.response:
		slog.Info("lobby.hostInstanceController received answer response", "isSuccess", connected)
		if !connected {
			return errors.New("pipe connection could not be established")
		}
	}
	return nil
}

func (c *hostInstanceController) onRemoteHostPipeConnectionRequest(offer *webrtc.SessionDescription, instanceId uuid.UUID) (*webrtc.SessionDescription, error) {
	ctx, cancelReq := context.WithCancel(context.Background())
	defer cancelReq()

	request := newLobbyRequest(ctx, instanceId)
	data := newHostGetPipeAnswerData(offer)
	request.data = data
	go c.lobby.runRequest(request)

	select {
	case <-c.ctx.Done():
		return nil, errors.New("cancel request while the main context is finished")
	case err := <-request.err:
		return nil, fmt.Errorf("getting host answer: %w", err)
	case res := <-data.response:
		slog.Info("lobby.hostInstanceController: getting host answer response success")
		return res.answer, nil
	}
}

func (c *hostInstanceController) onRemoteHostIngressConnectionRequest(offer *webrtc.SessionDescription, instanceId uuid.UUID) (*webrtc.SessionDescription, error) {
	ctx, cancelReq := context.WithCancel(context.Background())
	defer cancelReq()
	request := newLobbyRequest(ctx, instanceId)
	data := newHostGetIngressAnswerData(offer)
	request.data = data
	go c.lobby.runRequest(request)

	select {
	case <-c.ctx.Done():
		return nil, errors.New("cancel request while the main context is finished")
	case err := <-request.err:
		return nil, fmt.Errorf("getting host ingress answer: %w", err)
	case res := <-data.response:
		slog.Info("lobby.hostInstanceController: getting host ingress answer response success")

		if c.settings.isHost {
			go func() {
				remoteHost, _ := url.Parse("http://stream.localhost:8080")
				c.hostApi.url = remoteHost
				slog.Debug("lobby.hostInstanceController: start egress endpoint", "host", c.hostApi.url)
				if _, err := c.hostApi.Login(); err != nil {
					slog.Error("login to remote host", "err", err)
					return
				}
				if err := c.connectToHostEgress(instanceId); err != nil {
					slog.Error("lobby.hostInstanceController: connecting egress", "err", err)
					return
				}
			}()
		}

		return res.answer, nil
	}
}

type hostInstanceSettings struct {
	url        *url.URL
	space      string
	stream     string
	isHost     bool
	instanceId uuid.UUID
	name       string
	token      string
}
