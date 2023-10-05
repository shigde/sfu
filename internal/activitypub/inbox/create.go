package inbox

import (
	"context"

	"github.com/pkg/errors"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

type createInbox struct {
}

func newCreateInbox() *createInbox {
	return &createInbox{}
}

func (cr *createInbox) handleCreateRequest(ctx context.Context, activity vocab.ActivityStreamsCreate) error {
	iri := activity.GetJSONLDId().GetIRI().String()
	return errors.New("not handling create request of: " + iri)
}
