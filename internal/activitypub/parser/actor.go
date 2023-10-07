package parser

import (
	"errors"
	"net/url"

	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

func ExtractActorURI(withActor WithActor) (*url.URL, error) {
	actorProp := withActor.GetActivityStreamsActor()
	if actorProp == nil {
		return nil, errors.New("actor property was nil")
	}

	for iter := actorProp.Begin(); iter != actorProp.End(); iter = iter.Next() {
		id, err := pub.ToId(iter)
		if err == nil {
			// Found one we can use.
			return id, nil
		}
	}

	return nil, errors.New("no iri found for actor prop")
}

type WithActor interface {
	GetActivityStreamsActor() vocab.ActivityStreamsActorProperty
}
