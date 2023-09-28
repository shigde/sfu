package models

import (
	"net/url"
	"path"
)

func buildAccountIri(instanceUrl *url.URL, account string) *url.URL {
	iri, _ := url.Parse(path.Join(instanceUrl.Path, "federation", "account", account))
	return iri
}

func buildInboxIri(actorUrl *url.URL) *url.URL {
	iri, _ := url.Parse(path.Join(actorUrl.Path, "inbox"))
	return iri
}

func buildSharedInboxIri(instanceUrl *url.URL) *url.URL {
	iri, _ := url.Parse(path.Join(instanceUrl.Path, "federation", "inbox"))
	return iri
}

func buildOutboxIri(actorUrl *url.URL) *url.URL {
	iri, _ := url.Parse(path.Join(actorUrl.Path, "outbox"))
	return iri
}

func buildFollowersIri(actorUrl *url.URL) *url.URL {
	iri, _ := url.Parse(path.Join(actorUrl.Path, "followers"))
	return iri
}

func buildFollowingIri(actorUrl *url.URL) *url.URL {
	iri, _ := url.Parse(path.Join(actorUrl.Path, "following"))
	return iri
}
