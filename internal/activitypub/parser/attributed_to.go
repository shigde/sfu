package parser

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

func ExtractAttributedTo(withAttrTo WithAttributedTo) ([]*url.URL, error) {
	attrToProp := withAttrTo.GetActivityStreamsAttributedTo()
	if attrToProp == nil {
		return nil, errors.New("activity had no attributedTo")
	}
	toIris := make([]*url.URL, 0, attrToProp.Len())
	for iter := attrToProp.Begin(); iter != attrToProp.End(); iter = iter.Next() {
		actorURI, err := pub.ToId(iter)
		if err != nil {
			return nil, fmt.Errorf("error extracting id from attributedTo entry: %w", err)
		}
		// owner and channel
		toIris = append(toIris, actorURI)
	}

	return toIris, nil
}

type WithAttributedTo interface {
	GetActivityStreamsAttributedTo() vocab.ActivityStreamsAttributedToProperty
}
