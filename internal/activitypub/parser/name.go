package parser

import (
	"github.com/superseriousbusiness/activity/streams/vocab"
)

func ExtractName(withName WithName) string {
	name := withName.GetActivityStreamsName()
	return name.Name()
}

type WithName interface {
	GetActivityStreamsName() vocab.ActivityStreamsNameProperty
}
