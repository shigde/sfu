package inbox

import (
	"context"
	"errors"
	"fmt"

	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/activitypub/parser"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"golang.org/x/exp/slog"
)

type acceptInbox struct {
	followStore *models.FollowRepository
}

func newAcceptInbox(followStore *models.FollowRepository) *acceptInbox {
	return &acceptInbox{followStore: followStore}
}

/*
	{
		"type": "Accept",
		"id": "http://localhost:9000/accepts/follows/2",
		"actor": "http://localhost:9000/accounts/peertube",
		"object": {
			"type": "Follow",
			"id": "http://localhost:8080/federation/follow/38d45114-f79b-41be-9fbd-655ccf52e276",
			"actor": "http://localhost:8080/federation/accounts/shig",
			"object": "http://localhost:9000/accounts/peertube"
		}
	}
*/
func (ai *acceptInbox) handleAcceptRequest(ctx context.Context, activity vocab.ActivityStreamsAccept) error {
	slog.Debug("inbox accept: receive accept activity", "activity", activity)

	acceptObject := activity.GetActivityStreamsObject()
	if acceptObject == nil {
		return errors.New("inbox accept: no object set on vocab.ActivityStreamsAccept")
	}

	for iter := acceptObject.Begin(); iter != acceptObject.End(); iter = iter.Next() {

		if iter.IsIRI() {
			acceptedObjectIRI := iter.GetIRI()

			if instance.IsFollowActivityIri(acceptedObjectIRI) {
				// ACCEPT FOLLOW
				follow, err := ai.followStore.GetFollowByIri(ctx, acceptedObjectIRI.String())
				if err != nil {
					return fmt.Errorf("inbox accept: getting follow request with id %s from the database: %w", acceptedObjectIRI.String(), err)
				}

				//if follow.ActorId != receivingAccount.ID {
				//	return errors.New("inbox accept: follow object account and inbox account were not the same")
				//}

				if err = ai.saveFollowAsAccepted(ctx, follow); err != nil {
					return fmt.Errorf("saving follow in case one: %w", err)
				}
			}
		}

		if iter.GetType() == nil {
			continue
		}
		if iter.GetType().GetTypeName() == "Follow" {
			// ACCEPT FOLLOW
			asFollow, ok := iter.GetType().(vocab.ActivityStreamsFollow)
			if !ok {
				return errors.New("inbox accept: couldn't parse follow into vocab.ActivityStreamsFollow")
			}
			idProp := asFollow.GetJSONLDId()
			if idProp == nil || !idProp.IsIRI() {
				return errors.New("no id property set on follow, or was not an iri")
			}
			uri := idProp.GetIRI().String()
			follow, err := ai.followStore.GetFollowByIri(ctx, uri)
			if err != nil {
				return fmt.Errorf("inbox accept: getting follow request with id %s from the database: %w", uri, err)
			}

			if actorIri, err := parser.ExtractActorURI(asFollow); err != nil || actorIri.String() != follow.Actor.ActorIri {
				return fmt.Errorf("comparing actor with follow actvity: %w", err)
			}

			if targetIri, err := parser.ExtractObjectURI(asFollow); err != nil || targetIri.String() != follow.TargetActor.ActorIri {
				return fmt.Errorf("comparing target actor with follow actvity: %w", err)
			}

			if err = ai.saveFollowAsAccepted(ctx, follow); err != nil {
				return fmt.Errorf("saving follow in case two: %w", err)
			}
		}
	}

	return nil
}

func (ai *acceptInbox) saveFollowAsAccepted(ctx context.Context, follow *models.Follow) error {
	follow.State = models.Accepted.String()
	_, err := ai.followStore.Update(ctx, follow)
	if err != nil {
		return fmt.Errorf("inbox accept: updating follow request with id %s in the database: %w", follow.Iri, err)
	}
	return nil
}
