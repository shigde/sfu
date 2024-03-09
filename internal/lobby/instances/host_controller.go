package instances

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	lobby2 "github.com/shigde/sfu/internal/lobby"
	"github.com/shigde/sfu/internal/lobby/clients"
	"golang.org/x/exp/slog"
)

type hostController struct {
	ctx      context.Context
	lobbyId  uuid.UUID
	lobby    *lobby2.lobby
	hostApi  *clients.hostApiClient
	settings *hostSettings
}

func newHostController(ctx context.Context, lobbyId uuid.UUID, lobby *lobby2.lobby, settings *hostSettings) *hostController {
	hostApi := clients.newHostApiClient(settings.instanceId, settings.token, settings.name, settings.url)
	controller := &hostController{
		ctx:      ctx,
		lobbyId:  lobbyId,
		lobby:    lobby,
		hostApi:  hostApi,
		settings: settings,
	}

	go controller.run()
	if !settings.isHost {
		go func() {
			slog.Debug("lobby.hostController. connect to live stream host instance", "instanceId", settings.instanceId)
			if _, err := hostApi.Login(); err != nil {
				slog.Error("login to remote host", "err", err)
				return
			}

			if err := controller.connectToHostPipe(settings.instanceId); err != nil {
				slog.Error("lobby.hostController: connecting pipe", "err", err)
				return
			}
			if err := controller.connectToHostEgress(settings.instanceId); err != nil {
				slog.Error("lobby.hostController: connecting ingress", "err", err)
				return
			}
		}()
	}
	return controller
}

func (c *hostController) run() {
	slog.Info("lobby.hostController: run", "lobbyId", c.lobbyId)
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}
	}
}

func (c *hostController) connectToHostPipe(instanceId uuid.UUID) error {
	ctx, cancelReq := context.WithCancel(context.Background())
	defer cancelReq()

	request := lobby2.newLobbyRequest(ctx, instanceId)
	data := lobby2.newHostGetPipeOfferData()
	request.data = data
	go c.lobby.runRequest(request)

	var answer *webrtc.SessionDescription
	select {
	case <-c.ctx.Done():
		return lobby2.errSessionAlreadyClosed
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

func (c *hostController) onHostPipeAnswerResponse(answer *webrtc.SessionDescription, instanceId uuid.UUID) error {
	ctx, cancelReq := context.WithCancel(context.Background())
	defer cancelReq()
	request := lobby2.newLobbyRequest(ctx, instanceId)
	data := lobby2.newHostSetPipeAnswerData(answer)
	request.data = data
	go c.lobby.runRequest(request)

	select {
	case <-c.ctx.Done():
		return lobby2.errSessionAlreadyClosed
	case err := <-request.err:
		return fmt.Errorf("setting answer for pip connection: %w", err)
	case connected := <-data.response:
		slog.Info("lobby.hostController received answer response", "isSuccess", connected)
		if !connected {
			return errors.New("pipe connection could not be established")
		}
	}
	return nil
}

func (c *hostController) connectToHostEgress(instanceId uuid.UUID) error {
	ctx, cancelReq := context.WithCancel(context.Background())
	defer cancelReq()

	request := lobby2.newLobbyRequest(ctx, instanceId)
	data := lobby2.newHostGetEgressOfferData()
	request.data = data
	go c.lobby.runRequest(request)

	var answer *webrtc.SessionDescription
	select {
	case <-c.ctx.Done():
		return lobby2.errSessionAlreadyClosed
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

func (c *hostController) onHostEgressAnswerResponse(answer *webrtc.SessionDescription, instanceId uuid.UUID) error {
	ctx, cancelReq := context.WithCancel(context.Background())
	defer cancelReq()
	request := lobby2.newLobbyRequest(ctx, instanceId)
	data := lobby2.newHostSetEgressAnswerData(answer)
	request.data = data
	go c.lobby.runRequest(request)

	select {
	case <-c.ctx.Done():
		return lobby2.errSessionAlreadyClosed
	case err := <-request.err:
		return fmt.Errorf("setting answer for pip connection: %w", err)
	case connected := <-data.response:
		slog.Info("lobby.hostController received answer response", "isSuccess", connected)
		if !connected {
			return errors.New("pipe connection could not be established")
		}
	}
	return nil
}

func (c *hostController) onRemoteHostPipeConnectionRequest(offer *webrtc.SessionDescription, instanceId uuid.UUID) (*webrtc.SessionDescription, error) {
	ctx, cancelReq := context.WithCancel(context.Background())
	defer cancelReq()

	request := lobby2.newLobbyRequest(ctx, instanceId)
	data := lobby2.newHostGetPipeAnswerData(offer)
	request.data = data
	go c.lobby.runRequest(request)

	select {
	case <-c.ctx.Done():
		return nil, errors.New("cancel request while the main context is finished")
	case err := <-request.err:
		return nil, fmt.Errorf("getting host answer: %w", err)
	case res := <-data.response:
		slog.Info("lobby.hostController: getting host answer response success")
		return res.answer, nil
	}
}

func (c *hostController) onRemoteHostIngressConnectionRequest(offer *webrtc.SessionDescription, instanceId uuid.UUID) (*webrtc.SessionDescription, error) {
	ctx, cancelReq := context.WithCancel(context.Background())
	defer cancelReq()
	request := lobby2.newLobbyRequest(ctx, instanceId)
	data := lobby2.newHostGetIngressAnswerData(offer)
	request.data = data
	go c.lobby.runRequest(request)

	select {
	case <-c.ctx.Done():
		return nil, errors.New("cancel request while the main context is finished")
	case err := <-request.err:
		return nil, fmt.Errorf("getting host ingress answer: %w", err)
	case res := <-data.response:
		slog.Info("lobby.hostController: getting host ingress answer response success")

		if c.settings.isHost {
			go func() {
				remoteHost, _ := url.Parse("http://stream.localhost:8080")
				c.hostApi.url = remoteHost
				slog.Debug("lobby.hostController: start egress endpoint", "host", c.hostApi.url)
				if _, err := c.hostApi.Login(); err != nil {
					slog.Error("login to remote host", "err", err)
					return
				}
				if err := c.connectToHostEgress(instanceId); err != nil {
					slog.Error("lobby.hostController: connecting egress", "err", err)
					return
				}
			}()
		}

		return res.answer, nil
	}
}
