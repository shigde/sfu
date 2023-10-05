package inbox

import (
	"context"

	"github.com/shigde/sfu/internal/activitypub/remote"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"golang.org/x/exp/slog"
)

type updateInbox struct {
	resolver *remote.Resolver
}

func newUpdateInbox(resolver *remote.Resolver) *updateInbox {
	return &updateInbox{resolver: resolver}
}

func (u *updateInbox) handleUpdateRequest(ctx context.Context, activity vocab.ActivityStreamsUpdate) error {
	// We only care about update events to followers.
	if !activity.GetActivityStreamsObject().At(0).IsActivityStreamsPerson() {
		return nil
	}

	_, err := u.resolver.GetResolvedActorFromActorProperty(activity.GetActivityStreamsActor())
	if err != nil {
		slog.Error("err", err)
		return err
	}

	return nil
}
