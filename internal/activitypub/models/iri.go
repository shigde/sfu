package models

import (
	"net/url"
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
