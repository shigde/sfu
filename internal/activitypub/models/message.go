package models

import (
	"net/url"

	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

func CreateCreateActivity(id string, instanceIri *url.URL, localActorIRI *url.URL) vocab.ActivityStreamsCreate {
	objectID := instance.BuildResourceIri(instanceIri, id)
	message := MakeCreateActivity(objectID)

	actorProp := streams.NewActivityStreamsActorProperty()
	actorProp.AppendIRI(localActorIRI)
	message.SetActivityStreamsActor(actorProp)

	return message
}
