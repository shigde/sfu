package instances

import (
	"context"
	"net/url"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/lobby/clients"
	"golang.org/x/exp/slog"
)

type Connector struct {
	ctx          context.Context
	homeActorIri url.URL
	api          *clients.ApiClient
	host         *host
	space        string
	liveStream   string
}

func NewConnector(
	ctx context.Context,
	homeActorIri url.URL,
	hostActorIri url.URL,
	space string,
	liveStream string,
	token string,
) *Connector {
	host := newHost(hostActorIri, token)
	api := clients.NewApiClient(host, hostActorIri.Host)
	return &Connector{
		ctx:          ctx,
		homeActorIri: homeActorIri,
		api:          api,
		host:         host,
		space:        space,
		liveStream:   liveStream,
	}
}

func (c *Connector) Connect() {
	slog.Debug("lobby.HostController. connect to live stream host instance", "instanceId", c.host.instanceId)
	if _, err := c.api.Login(); err != nil {
		slog.Error("login to remote host", "err", err)
		return
	}

	if err := c.connectIngress(c.host.instanceId); err != nil {
		slog.Error("lobby.HostController: connecting pipe", "err", err)
		return
	}
	if err := c.connectEgress(c.host.instanceId); err != nil {
		slog.Error("lobby.HostController: connecting ingress", "err", err)
		return
	}
}

func (c *Connector) connectIngress(id uuid.UUID) error {
	// ctx, cancelReq := context.WithCancel(context.Background())
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
	return nil
}

func (c *Connector) connectEgress(id uuid.UUID) interface{} {
	return nil
}
