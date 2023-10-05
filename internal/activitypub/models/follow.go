package models

import (
	"fmt"
	"net/url"

	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"gorm.io/gorm"
)

type FollowState uint

const (
	Accepted FollowState = iota
	Pending
	Rejected
)

func (fs FollowState) String() string {
	return []string{"accepted", "pending", "rejected"}[fs]
}

type Follow struct {
	Iri           string   `gorm:"iri;not null;index;"`
	ActorId       uint     `gorm:"actor_id;not null;"`
	TargetActorId uint     `gorm:"target_actor_id;not null;"`
	State         string   `gorm:"state;not null;"`
	Actor         *Actor   `gorm:"foreignKey:ActorId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	TargetActor   *Actor   `gorm:"foreignKey:TargetActorId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	iriUrl        *url.URL `gorm:"-"`
	gorm.Model
}

func NewFollow(actor *Actor, target *Actor, config *instance.FederationConfig) *Follow {
	iri := instance.BuildFollowActivityIri(config.InstanceUrl)
	return &Follow{
		Iri:           iri.String(),
		ActorId:       actor.ID,
		TargetActorId: target.ID,
		Actor:         actor,
		TargetActor:   target,
		State:         Pending.String(),
		iriUrl:        iri,
	}
}
func (f *Follow) GetIri() *url.URL {
	if f.iriUrl == nil {
		f.iriUrl, _ = url.Parse(f.Iri)
	}
	return f.iriUrl
}

func (f *Follow) ToAS() (vocab.ActivityStreamsFollow, error) {
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
	followURI, err := url.Parse(f.Iri)
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
	//followToProp := streams.NewActivityStreamsToProperty()
	//followToProp.AppendIRI(targetAccountURI)
	//follow.SetActivityStreamsTo(followToProp)

	return follow, nil
}
