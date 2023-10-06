package inbox

import (
	"context"
	"errors"
	"fmt"

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
	asObject := activity.GetActivityStreamsObject()
	if asObject == nil {
		return errors.New("no object set on vocab.ActivityStreamsCreate")
	}

	if err := cr.videoService.UpsertVideo(ctx, asObject); err != nil {
		return fmt.Errorf("inbox update video: %w", err)
	}

	return nil
}
