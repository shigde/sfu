package inbox

import (
	"context"

	"github.com/pkg/errors"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

func handleCreateRequest(c context.Context, activity vocab.ActivityStreamsCreate) error {
	iri := activity.GetJSONLDId().GetIRI().String()
	return errors.New("not handling create request of: " + iri)
}
