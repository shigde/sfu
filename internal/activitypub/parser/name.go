package parser

import (
	"github.com/superseriousbusiness/activity/streams/vocab"
)

func ExtractName(withName WithName) string {
	if withName.GetActivityStreamsName().Len() == 1 {
		return withName.GetActivityStreamsName().At(0).GetXMLSchemaString()
	}
	return ""
}

type WithName interface {
	GetActivityStreamsName() vocab.ActivityStreamsNameProperty
}
