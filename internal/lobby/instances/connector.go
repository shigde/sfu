package instances

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/lobby/clients"
	"github.com/shigde/sfu/internal/lobby/commands"
	"github.com/shigde/sfu/internal/lobby/sessions"
	"golang.org/x/exp/slog"
)

var (
	LoginError = errors.New("login to remote instance failed")
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

type command interface {
	GetUserId() uuid.UUID
	Execute(session *sessions.Session)
	SetError(err error)
}

func (c *Connector) BuildIngress() (*commands.Command, error) {
	slog.Debug("lobby.HostController. connect to live stream host instance", "instanceId", c.host.instanceId)
	if _, err := c.api.Login(); err != nil {
		return nil, fmt.Errorf("login to remote host: %w", err)
	}
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
	return nil, nil
}

func (c *Connector) BuildEgress() (*commands.Command, error) {
	return nil, nil
}

func (c *Connector) IsThisInstanceLiveSteamHost() bool {
	return c.homeActorIri.String() == c.host.actorIri.String()
}

func (c *Connector) GetInstanceId() uuid.UUID {
	return c.host.instanceId
}
