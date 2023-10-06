package inbox

import (
	"context"

	"github.com/pkg/errors"
	"github.com/shigde/sfu/internal/activitypub/services"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

type createInbox struct {
	videoService *services.VideoService
}

func newCreateInbox(videoService *services.VideoService) *createInbox {
	return &createInbox{
		videoService: videoService,
	}
}

func (cr *createInbox) handleCreateRequest(ctx context.Context, activity vocab.ActivityStreamsCreate) error {
	// We only care about video creates.
	if !activity.GetActivityStreamsObject().At(0).IsActivityStreamsVideo() {
		return nil
	}

	iri := activity.GetJSONLDId().GetIRI().String()
	return errors.New("not handling create request of: " + iri)
}
