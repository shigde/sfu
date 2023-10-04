package models

import (
	"context"
	"fmt"
	"net/url"

	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

type Follow struct {
	Iri         *url.URL
	Actor       *Actor
	TargetActor *Actor
}

func NewFollow(actor *Actor, target *Actor, config *instance.FederationConfig) *Follow {
	iri := instance.BuildFollowActivityIri(config.InstanceUrl)
	return &Follow{
		Iri:         iri,
		Actor:       actor,
		TargetActor: target,
	}
}

/**
{
"@context":"https://www.w3.org/ns/activitystreams",
"id":"https://activitypub.academy/16606771-befe-483b-9e2c-0b8b85062373",
"type":"Follow",
"actor":"https://activitypub.academy/users/alice",
"object":"https://techhub.social/users/berta"
}
*/

func (f *Follow) ToAS(ctx context.Context) (vocab.ActivityStreamsFollow, error) {
	//if err := c.state.DB.PopulateFollow(ctx, f); err != nil {
	//	return nil, gtserror.Newf("error populating follow: %w", err)
	//}

	// Parse out the various URIs we need for this
	// origin account (who's doing the follow).
	originAccountURI, err := url.Parse(f.Actor.ActorIri)
	if err != nil {
		return nil, fmt.Errorf("followtoasfollow: error parsing origin account uri: %s", err)
	}
	originActor := streams.NewActivityStreamsActorProperty()
	originActor.AppendIRI(originAccountURI)

	// target account (who's being followed)
	targetAccountURI, err := url.Parse(f.TargetActor.ActorIri)
	if err != nil {
		return nil, fmt.Errorf("followtoasfollow: error parsing target account uri: %s", err)
	}

	// uri of the follow activity itself
	followURI, err := url.Parse(f.Iri.Path)
	if err != nil {
		return nil, fmt.Errorf("followtoasfollow: error parsing follow uri: %s", err)
	}

	// start preparing the follow activity
	follow := streams.NewActivityStreamsFollow()

	// set the actor
	follow.SetActivityStreamsActor(originActor)

	// set the id
	followIDProp := streams.NewJSONLDIdProperty()
	followIDProp.SetIRI(followURI)
	follow.SetJSONLDId(followIDProp)

	// set the object
	followObjectProp := streams.NewActivityStreamsObjectProperty()
	followObjectProp.AppendIRI(targetAccountURI)
	follow.SetActivityStreamsObject(followObjectProp)

	// set the To property
	followToProp := streams.NewActivityStreamsToProperty()
	followToProp.AppendIRI(targetAccountURI)
	follow.SetActivityStreamsTo(followToProp)

	return follow, nil
}
