package inbox

import (
	"context"
	"errors"
	"fmt"

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
	asObject := activity.GetActivityStreamsObject()
	if asObject == nil {
		return errors.New("no object set on vocab.ActivityStreamsUpdate")
	}

	if err := u.videoService.UpsertVideo(ctx, asObject); err != nil {
		return fmt.Errorf("inbox update video: %w", err)
	}

	return nil
}
