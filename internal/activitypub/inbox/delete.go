package inbox

import (
	"context"
	"errors"
	"fmt"

	"github.com/shigde/sfu/internal/activitypub/services"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

type deleteInbox struct {
	videoService *services.VideoService
}

func newDeleteInbox(videoService *services.VideoService) *deleteInbox {
	return &deleteInbox{
		videoService: videoService,
	}
}

func (u *deleteInbox) handleDeleteRequest(ctx context.Context, activity vocab.ActivityStreamsDelete) error {
	// We only care about video updates.
	asObject := activity.GetActivityStreamsObject()
	if asObject == nil {
		return errors.New("no object set on vocab.ActivityStreamsCreate")
	}

	if err := u.videoService.DeleteVideo(ctx, asObject); err != nil {
		return fmt.Errorf("inbox update video: %w", err)
	}

	return nil
}
