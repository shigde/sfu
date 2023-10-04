package inbox

import (
	"context"

	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/activitypub/remote"
	log "github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

type updateInbox struct {
	resolver      *remote.Resolver
	followerStore *models.FollowRepository
}

func (u *updateInbox) handleUpdateRequest(ctx context.Context, activity vocab.ActivityStreamsUpdate) error {
	// We only care about update events to followers.
	if !activity.GetActivityStreamsObject().At(0).IsActivityStreamsPerson() {
		return nil
	}

	actor, err := u.resolver.GetResolvedActorFromActorProperty(activity.GetActivityStreamsActor())
	if err != nil {
		log.Errorln(err)
		return err
	}

	return u.followerStore.UpdateFollower(ctx, &models.ActorFollow{ActorId: actor.ID})
}
