package inbox

import (
	"context"
	"errors"
	"fmt"

	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/activitypub/models"
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

				follow.State = models.Accepted.String()
				_, err = ai.followStore.Update(ctx, follow)
				if err != nil {
					return fmt.Errorf("inbox accept: updating follow request with id %s in the database: %w", acceptedObjectIRI.String(), err)
				}
			}

		}
	}

	return nil
}

func (ai *acceptInbox) Accept(ctx context.Context, accept vocab.ActivityStreamsAccept) error {
	//if log.Level() >= level.DEBUG {
	//	i, err := marshalItem(accept)
	//	if err != nil {
	//		return err
	//	}
	//	l := log.WithContext(ctx).
	//		WithField("accept", i)
	//	l.Debug("entering Accept")
	//}

	//receivingAccount, _, internal := extractFromCtx(ctx)
	//if internal {
	//	return nil // Already processed.
	//}

	acceptObject := accept.GetActivityStreamsObject()
	if acceptObject == nil {
		return errors.New("inbox accept: no object set on vocab.ActivityStreamsAccept")
	}
	//
	//for iter := acceptObject.Begin(); iter != acceptObject.End(); iter = iter.Next() {
	//	// check if the object is an IRI
	//	if iter.IsIRI() {
	//		// we have just the URI of whatever is being accepted, so we need to find out what it is
	//		acceptedObjectIRI := iter.GetIRI()
	//
	//		ai.followStore.g
	//		if uris.IsFollowPath(acceptedObjectIRI) {
	//			// ACCEPT FOLLOW
	//			followReq, err := f.state.DB.GetFollowRequestByURI(ctx, acceptedObjectIRI.String())
	//			if err != nil {
	//				return fmt.Errorf("ACCEPT: couldn't get follow request with id %s from the database: %s", acceptedObjectIRI.String(), err)
	//			}
	//
	//			// make sure the addressee of the original follow is the same as whatever inbox this landed in
	//			if followReq.AccountID != receivingAccount.ID {
	//				return errors.New("ACCEPT: follow object account and inbox account were not the same")
	//			}
	//			follow, err := f.state.DB.AcceptFollowRequest(ctx, followReq.AccountID, followReq.TargetAccountID)
	//			if err != nil {
	//				return err
	//			}
	//
	//			f.state.Workers.EnqueueFediAPI(ctx, messages.FromFediAPI{
	//				APObjectType:     ap.ActivityFollow,
	//				APActivityType:   ap.ActivityAccept,
	//				GTSModel:         follow,
	//				ReceivingAccount: receivingAccount,
	//			})
	//
	//			return nil
	//		}
	//	}
	//
	//	// check if iter is an AP object / type
	//	if iter.GetType() == nil {
	//		continue
	//	}
	//	if iter.GetType().GetTypeName() == ap.ActivityFollow {
	//		// ACCEPT FOLLOW
	//		asFollow, ok := iter.GetType().(vocab.ActivityStreamsFollow)
	//		if !ok {
	//			return errors.New("ACCEPT: couldn't parse follow into vocab.ActivityStreamsFollow")
	//		}
	//		// convert the follow to something we can understand
	//		gtsFollow, err := f.converter.ASFollowToFollow(ctx, asFollow)
	//		if err != nil {
	//			return fmt.Errorf("ACCEPT: error converting asfollow to gtsfollow: %s", err)
	//		}
	//		// make sure the addressee of the original follow is the same as whatever inbox this landed in
	//		if gtsFollow.AccountID != receivingAccount.ID {
	//			return errors.New("ACCEPT: follow object account and inbox account were not the same")
	//		}
	//		follow, err := f.state.DB.AcceptFollowRequest(ctx, gtsFollow.AccountID, gtsFollow.TargetAccountID)
	//		if err != nil {
	//			return err
	//		}
	//
	//		f.state.Workers.EnqueueFediAPI(ctx, messages.FromFediAPI{
	//			APObjectType:     ap.ActivityFollow,
	//			APActivityType:   ap.ActivityAccept,
	//			GTSModel:         follow,
	//			ReceivingAccount: receivingAccount,
	//		})
	//
	//		return nil
	//	}
	//}

	return nil
}
