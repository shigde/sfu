package parser

import (
	"errors"
	"net/url"

	"github.com/superseriousbusiness/activity/streams/vocab"
)

func ExtractTo(withAttrTo WithTo) ([]*url.URL, error) {
	attrTo := withAttrTo.GetActivityStreamsTo()
	if attrTo == nil {
		return nil, errors.New("activity had no attributedTo")
	}
	toIris := make([]*url.URL, 0, attrTo.Len())

	for iter := attrTo.Begin(); iter != attrTo.End(); iter = iter.Next() {
		actorURI := iter.GetIRI()
		toIris = append(toIris, actorURI)
	}

	return toIris, nil
}

type WithTo interface {
	GetActivityStreamsTo() vocab.ActivityStreamsToProperty
}
