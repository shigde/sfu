package inbox

import (
	"context"
	"errors"
	"fmt"

	"github.com/shigde/sfu/internal/activitypub/parser"
	"github.com/shigde/sfu/internal/activitypub/services"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

type announceInbox struct {
	videoService *services.VideoService
}

func newAnnounceInbox(videoService *services.VideoService) *announceInbox {
	return &announceInbox{
		videoService: videoService,
	}
}

func (cr *announceInbox) handleAnnounceRequest(ctx context.Context, activity vocab.ActivityStreamsAnnounce) error {
	// We only care about video creates.

	asObject := activity.GetActivityStreamsObject()
	if asObject == nil {
		return errors.New("no object set on vocab.ActivityStreamsCreate")
	}
	toFollowerIris, err := parser.ExtractTo(activity)
	if err != nil {
		return errors.New("no object set on vocab.ActivityStreamsCreate")
	}

	if err := cr.videoService.AddVideo(ctx, asObject, toFollowerIris); err != nil {
		return fmt.Errorf("inbox announce video: %w", err)
	}

	return nil
}
