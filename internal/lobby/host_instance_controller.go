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
	hostApi := NewHostInstanceApiClient(settings.instanceId, settings.token, settings.url)
	controller := &hostInstanceController{
		ctx:      ctx,
		lobbyId:  lobbyId,
		lobby:    lobby,
		hostApi:  hostApi,
		settings: settings,
	}

	go controller.run()
	if !settings.isHost {
		slog.Debug("lobby.hostInstanceController. connect to live stream host instance", "instanceId", settings.instanceId)
		controller.connectToHost(settings.instanceId)
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

func (c *hostInstanceController) connectToHost(instanceId uuid.UUID) {
	ctx, cancelReq := context.WithCancel(context.Background())
	request := newLobbyRequest(ctx, c.lobbyId)
	data := newHostGetOfferData()
	request.data = data
	go func() {
		select {
		case <-c.ctx.Done():
			cancelReq()
		case err := <-request.err:
			slog.Error("connect to host", "err", fmt.Errorf("requesting joining lobby: %w", err))
		case res := <-data.response:
			if answer, err := c.hostApi.PostHostOffer(c.settings.space, c.settings.stream, res.offer); err == nil {
				c.onHostAnswerResponse(answer, instanceId)
			} else {
				slog.Error("lobby.hostInstanceController: remote host answer request", "err", err)
			}
		}
	}()
	go c.lobby.runRequest(request)
}

func (c *hostInstanceController) onHostAnswerResponse(answer *webrtc.SessionDescription, instanceId uuid.UUID) {
	ctx, cancelReq := context.WithCancel(context.Background())
	request := newLobbyRequest(ctx, instanceId)
	data := newHostSetAnswerData(answer)
	request.data = data
	go func() {
		defer cancelReq()
		select {
		case <-c.ctx.Done():
		case err := <-request.err:
			slog.Error("lobby.hostInstanceController: connect to host", "err", fmt.Errorf("requesting joining lobby: %w", err))
		case connected := <-data.response:
			if !connected {
				slog.Error("lobby.hostInstanceController", "err", "connection could not be established!")
			}
		}
	}()
	go c.lobby.runRequest(request)
}

func (c *hostInstanceController) onHostConnectionRequest(offer *webrtc.SessionDescription, instanceId uuid.UUID) (*webrtc.SessionDescription, error) {
	ctx, cancelReq := context.WithCancel(context.Background())
	defer cancelReq()
	request := newLobbyRequest(ctx, instanceId)
	data := newHostGetAnswerData(offer)
	request.data = data
	go c.lobby.runRequest(request)

	select {
	case <-c.ctx.Done():
		return nil, errors.New("cancel request while the main context is finished")
	case err := <-request.err:
		return nil, fmt.Errorf("getting host answer: %w", err)
	case res := <-data.response:
		return res.answer, nil
	}
}

type hostInstanceSettings struct {
	url        *url.URL
	space      string
	stream     string
	isHost     bool
	instanceId uuid.UUID
	token      string
}
