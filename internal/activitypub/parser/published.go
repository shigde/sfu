package parser

import (
	"time"

	"github.com/superseriousbusiness/activity/streams/vocab"
)

func ExtractPublished(withPublished WithPublished) time.Time {
	published := withPublished.GetActivityStreamsPublished()
	return published.Get()
}

type WithPublished interface {
	GetActivityStreamsPublished() vocab.ActivityStreamsPublishedProperty
}
