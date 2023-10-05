package activitypub

import (
	"context"
	"net/http"
	"net/url"

	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

type CommonBehavior struct {
}

func NewCommonBehavior() *CommonBehavior {
	return &CommonBehavior{}
}

func (b *CommonBehavior) AuthenticateGetInbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	return c, true, nil
}

func (b *CommonBehavior) AuthenticateGetOutbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	return nil, false, nil
}

func (b *CommonBehavior) GetOutbox(c context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	return nil, nil
}

func (b *CommonBehavior) NewTransport(c context.Context, actorBoxIRI *url.URL, gofedAgent string) (t pub.Transport, err error) {
	return nil, nil
}
