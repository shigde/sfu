package parser

import (
	"errors"
	"net/url"

	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

func ExtractObjectURI(withObject WithObject) (*url.URL, error) {
	objectProp := withObject.GetActivityStreamsObject()
	if objectProp == nil {
		return nil, errors.New("object property was nil")
	}

	for iter := objectProp.Begin(); iter != objectProp.End(); iter = iter.Next() {
		id, err := pub.ToId(iter)
		if err == nil {
			// Found one we can use.
			return id, nil
		}
	}

	return nil, errors.New("no iri found for object prop")
}

type WithObject interface {
	GetActivityStreamsObject() vocab.ActivityStreamsObjectProperty
}
