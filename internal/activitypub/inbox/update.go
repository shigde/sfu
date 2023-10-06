package inbox

import (
	"context"

	"github.com/shigde/sfu/internal/activitypub/services"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

type updateInbox struct {
	videoService *services.VideoService
}

func newUpdateInbox(videoService *services.VideoService) *updateInbox {
	return &updateInbox{
		videoService: videoService,
	}
}

func (u *updateInbox) handleUpdateRequest(ctx context.Context, activity vocab.ActivityStreamsUpdate) error {
	// We only care about video updates.
	if !activity.GetActivityStreamsObject().At(0).IsActivityStreamsVideo() {
		return nil
	}

	return nil
}
