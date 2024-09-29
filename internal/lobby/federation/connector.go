package federation

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
	api := clients.NewApiClient(host, hostActorIri.Host, space, liveStream)
	return &Connector{
		ctx:          ctx,
		homeActorIri: homeActorIri,
		api:          api,
		host:         host,
		space:        space,
		liveStream:   liveStream,
	}
}

func (c *Connector) Login() error {
	slog.Debug("federation connector login to live stream host instance", "instanceId", c.host.instanceId)
	if _, err := c.api.Login(); err != nil {
		return fmt.Errorf("login to remote host: %w", err)
	}
	return nil
}

func (c *Connector) BuildIngress() (*commands.OfferIngress, error) {
	slog.Debug("federation connector build ingress for live stream host instance", "instanceId", c.host.instanceId)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := commands.NewOfferIngress(ctx, c.api, c.host.instanceId, sessions.BidirectionalSignalChannel)

	return cmd, nil
}

func (c *Connector) BuildEgress() (*commands.OfferEgress, error) {
	slog.Debug("federation connector build egress for live stream host instance", "instanceId", c.host.instanceId)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := commands.NewOfferEgress(ctx, c.api, c.host.instanceId, sessions.BidirectionalSignalChannel)

	return cmd, nil
}

func (c *Connector) IsThisInstanceLiveSteamHost() bool {
	return c.homeActorIri.String() == c.host.actorIri.String()
}

func (c *Connector) GetInstanceId() uuid.UUID {
	return c.host.instanceId
}
