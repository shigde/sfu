package models

import (
	"context"
	"net/url"

	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

func buildAccountIri(instanceUrl *url.URL, account string) *url.URL {
	iri, _ := url.Parse(instanceUrl.JoinPath("federation", "account", account).String())
	return iri
}

func buildInboxIri(actorUrl *url.URL) *url.URL {
	iri, _ := url.Parse(actorUrl.JoinPath("inbox").String())
	return iri
}

func buildSharedInboxIri(instanceUrl *url.URL) *url.URL {
	iri, _ := url.Parse(instanceUrl.JoinPath("federation", "inbox").String())
	return iri
}

func buildOutboxIri(actorUrl *url.URL) *url.URL {
	iri, _ := url.Parse(actorUrl.JoinPath("outbox").String())
	return iri
}

func buildFollowersIri(actorUrl *url.URL) *url.URL {
	iri, _ := url.Parse(actorUrl.JoinPath("followers").String())
	return iri
}

func buildFollowingIri(actorUrl *url.URL) *url.URL {
	iri, _ := url.Parse(actorUrl.JoinPath("following").String())
	return iri
}

func CollectIRIs(ctx context.Context, iris []*url.URL) (vocab.ActivityStreamsCollection, error) {
	collection := streams.NewActivityStreamsCollection()
	items := streams.NewActivityStreamsItemsProperty()
	for _, i := range iris {
		items.AppendIRI(i)
	}
	collection.SetActivityStreamsItems(items)
	return collection, nil
}
