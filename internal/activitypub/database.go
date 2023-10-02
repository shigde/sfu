package activitypub

// https://github.com/go-fed/testsuite/blob/master/server/db.go#L28

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"codeberg.org/gruf/go-mutexes"
	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

type Database struct {
	locks     mutexes.MutexMap
	actorPep  *models.ActorRepository
	followRep *models.ActorFollowRepository
}

func NewDatabase(actorPep *models.ActorRepository, followRep *models.ActorFollowRepository) *Database {
	return &Database{
		locks:     mutexes.NewMap(-1, -1),
		actorPep:  actorPep,
		followRep: followRep,
	}
}

func (d *Database) Lock(_ context.Context, id *url.URL) (unlock func(), err error) {
	if id == nil {
		return nil, errors.New("lock id was nil")
	}
	unlock = d.locks.Lock(id.String())
	return unlock, nil
}

func (d *Database) InboxContains(c context.Context, inbox, id *url.URL) (contains bool, err error) {
	return false, nil
}

func (d *Database) GetInbox(c context.Context, inboxIRI *url.URL) (inbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	return nil, nil
}

func (d *Database) SetInbox(c context.Context, inbox vocab.ActivityStreamsOrderedCollectionPage) error {
	return nil
}

func (d *Database) Owns(c context.Context, id *url.URL) (owns bool, err error) {
	return false, err
}

func (d *Database) ActorForOutbox(c context.Context, outboxIRI *url.URL) (actorIRI *url.URL, err error) {
	return nil, nil
}

func (d *Database) ActorForInbox(c context.Context, inboxIRI *url.URL) (actorIRI *url.URL, err error) {
	return nil, nil
}

func (d *Database) InboxesForIRI(c context.Context, iri *url.URL) (inboxIRIs []*url.URL, err error) {
	return nil, nil
}

// --- CRUD
func (d *Database) Exists(c context.Context, id *url.URL) (exists bool, err error) {
	return false, err
}

func (d *Database) Get(c context.Context, id *url.URL) (value vocab.Type, err error) {
	return nil, nil
}

func (d *Database) Create(c context.Context, asType vocab.Type) error {
	return nil
}

func (d *Database) Update(c context.Context, asType vocab.Type) error {
	return nil
}

func (d *Database) Delete(c context.Context, id *url.URL) error {
	return nil
}

// --- Outbox

// GetOutbox Implementation note: we don't (yet) serve outboxes, so just return empty and nil here.
func (d *Database) GetOutbox(c context.Context, outboxIRI *url.URL) (outbox vocab.ActivityStreamsOrderedCollectionPage, err error) {
	return streams.NewActivityStreamsOrderedCollectionPage(), nil
}

// SetOutbox Implementation note: we don't allow outbox setting so just return nil here.
func (d *Database) SetOutbox(c context.Context, outbox vocab.ActivityStreamsOrderedCollectionPage) error {
	return nil
}

func (d *Database) OutboxForInbox(ctx context.Context, inboxIRI *url.URL) (outboxIRI *url.URL, err error) {
	acct, err := d.actorPep.GetActorForIRI(ctx, inboxIRI, models.InboxIri)
	if err != nil {
		return nil, err
	}
	return url.Parse(acct.OutboxIri)
}

func (d *Database) NewID(c context.Context, t vocab.Type) (id *url.URL, err error) {
	return nil, nil
}

// --- Follow

func (d *Database) Followers(ctx context.Context, actorIRI *url.URL) (followers vocab.ActivityStreamsCollection, err error) {
	acct, err := d.actorPep.GetActorForIRI(ctx, actorIRI, models.ActorIri)
	if err != nil {
		return nil, err
	}

	// Fetch followers for account from database.
	follows, err := d.followRep.GetActorFollowers(ctx, acct.ID)
	if err != nil {
		return nil, fmt.Errorf("getting followers for actor id %d: %s", acct.ID, err)
	}
	actorIds := make([]uint, 0, len(follows))
	for _, follow := range follows {
		actorIds = append(actorIds, follow.ActorId)
	}

	actors, err := d.actorPep.GetAllActorsByIds(ctx, actorIds)
	if err != nil {
		return nil, fmt.Errorf("getting follower actors for actor ids %d: %w", acct.ID, err)
	}

	// Convert the followers to a slice of account URIs.
	iris := make([]*url.URL, 0, len(actors))
	for _, actor := range actors {
		u, err := url.Parse(actor.ActorIri)
		if err != nil {
			return nil, fmt.Errorf("parsing invalid account uri: %w", err)
		}
		iris = append(iris, u)
	}

	return instance.CollectIRIs(ctx, iris)
}

func (d *Database) Following(ctx context.Context, actorIRI *url.URL) (following vocab.ActivityStreamsCollection, err error) {
	acct, err := d.actorPep.GetActorForIRI(ctx, actorIRI, models.ActorIri)
	if err != nil {
		return nil, err
	}

	// Fetch follows for account from database.
	follows := acct.ActorFollow

	// Convert the follows to a slice of account URIs.
	iris := make([]*url.URL, 0, len(follows))
	for _, follow := range follows {
		u, err := url.Parse(follow.Iri)
		if err != nil {
			return nil, fmt.Errorf("parsing invalid account iri: %w, err")
		}
		iris = append(iris, u)
	}

	return instance.CollectIRIs(ctx, iris)

}

func (d *Database) Liked(c context.Context, actorIRI *url.URL) (liked vocab.ActivityStreamsCollection, err error) {
	return nil, nil
}
