package activitypub

import (
	"context"
	"net/http"
	"net/url"

	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

type FederatingProtocol struct {
}

func NewFederatingProtocol() *FederatingProtocol {
	return &FederatingProtocol{}
}

func (p *FederatingProtocol) PostInboxRequestBodyHook(c context.Context, r *http.Request, activity pub.Activity) (context.Context, error) {
	return c, nil
}

func (p *FederatingProtocol) AuthenticatePostInbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error) {
	return c, true, nil
}

func (p *FederatingProtocol) Blocked(c context.Context, actorIRIs []*url.URL) (blocked bool, err error) {
	return false, nil
}

func (p *FederatingProtocol) FederatingCallbacks(c context.Context) (wrapped pub.FederatingWrappedCallbacks, other []interface{}, err error) {
	return wrapped, other, nil
}

func (p *FederatingProtocol) DefaultCallback(c context.Context, activity pub.Activity) error {
	return nil
}

func (p *FederatingProtocol) MaxInboxForwardingRecursionDepth(_ context.Context) int {
	return 0
}

func (p *FederatingProtocol) MaxDeliveryRecursionDepth(_ context.Context) int {
	return 0
}

func (p *FederatingProtocol) FilterForwarding(c context.Context, potentialRecipients []*url.URL, a pub.Activity) (filteredRecipients []*url.URL, err error) {
	return filteredRecipients, nil
}

func (p *FederatingProtocol) GetInbox(c context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error) {
	return nil, nil
}
