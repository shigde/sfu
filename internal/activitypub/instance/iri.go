package instance

import (
	"context"
	"net/url"

	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

func BuildAccountIri(instanceUrl *url.URL, account string) *url.URL {
	iri, _ := url.Parse(instanceUrl.JoinPath("federation", "account", account).String())
	return iri
}

func BuildInboxIri(actorUrl *url.URL) *url.URL {
	iri, _ := url.Parse(actorUrl.JoinPath("inbox").String())
	return iri
}

func BuildSharedInboxIri(instanceUrl *url.URL) *url.URL {
	iri, _ := url.Parse(instanceUrl.JoinPath("federation", "inbox").String())
	return iri
}

func BuildOutboxIri(actorUrl *url.URL) *url.URL {
	iri, _ := url.Parse(actorUrl.JoinPath("outbox").String())
	return iri
}

func BuildFollowersIri(actorUrl *url.URL) *url.URL {
	iri, _ := url.Parse(actorUrl.JoinPath("followers").String())
	return iri
}

func BuildFollowingIri(actorUrl *url.URL) *url.URL {
	iri, _ := url.Parse(actorUrl.JoinPath("following").String())
	return iri
}

func BuildStreamURLIri(instanceUrl *url.URL) *url.URL {
	iri, _ := url.Parse(instanceUrl.JoinPath("hls", "stream.m3u8").String())
	return iri
}

func BuildResourceIri(instanceUrl *url.URL, resourcePath string) *url.URL {
	iri, _ := url.Parse(instanceUrl.JoinPath("federation", resourcePath).String())
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
