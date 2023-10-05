package models

import (
	"github.com/shigde/sfu/internal/activitypub/crypto"
	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"golang.org/x/exp/slog"
)

func BuildActivityApplication(actor *Actor, config *instance.FederationConfig) vocab.ActivityStreamsApplication {
	actorIRI := actor.GetActorIri()

	app := streams.NewActivityStreamsApplication()
	nameProperty := streams.NewActivityStreamsNameProperty()
	nameProperty.AppendXMLSchemaString(config.ServerName)
	app.SetActivityStreamsName(nameProperty)

	preferredUsernameProperty := streams.NewActivityStreamsPreferredUsernameProperty()
	preferredUsernameProperty.SetXMLSchemaString(actor.PreferredUsername)
	app.SetActivityStreamsPreferredUsername(preferredUsernameProperty)

	inboxProp := streams.NewActivityStreamsInboxProperty()
	inboxProp.SetIRI(actor.GetInboxIri())
	app.SetActivityStreamsInbox(inboxProp)

	needsFollowApprovalProperty := streams.NewActivityStreamsManuallyApprovesFollowersProperty()
	needsFollowApprovalProperty.Set(config.IsPrivate)
	app.SetActivityStreamsManuallyApprovesFollowers(needsFollowApprovalProperty)

	outboxProp := streams.NewActivityStreamsOutboxProperty()
	outboxProp.SetIRI(actor.GetOutboxIri())
	app.SetActivityStreamsOutbox(outboxProp)

	id := streams.NewJSONLDIdProperty()
	id.Set(actorIRI)
	app.SetJSONLDId(id)

	publicKey := crypto.GetPublicKey(actorIRI, actor.PublicKey)

	publicKeyProp := streams.NewW3IDSecurityV1PublicKeyProperty()
	publicKeyType := streams.NewW3IDSecurityV1PublicKey()

	pubKeyIDProp := streams.NewJSONLDIdProperty()
	pubKeyIDProp.Set(publicKey.ID)

	publicKeyType.SetJSONLDId(pubKeyIDProp)

	ownerProp := streams.NewW3IDSecurityV1OwnerProperty()
	ownerProp.SetIRI(publicKey.Owner)
	publicKeyType.SetW3IDSecurityV1Owner(ownerProp)

	publicKeyPemProp := streams.NewW3IDSecurityV1PublicKeyPemProperty()
	publicKeyPemProp.Set(publicKey.PublicKeyPem)
	publicKeyType.SetW3IDSecurityV1PublicKeyPem(publicKeyPemProp)
	publicKeyProp.AppendW3IDSecurityV1PublicKey(publicKeyType)
	app.SetW3IDSecurityV1PublicKey(publicKeyProp)

	if config.ServerInitTime.Valid {
		publishedDateProp := streams.NewActivityStreamsPublishedProperty()
		publishedDateProp.Set(config.ServerInitTime.Time)
		app.SetActivityStreamsPublished(publishedDateProp)
	} else {
		slog.Error("unable to fetch server init time")
	}

	// Profile properties

	// Avatar
	//uniquenessString := data.GetLogoUniquenessString()
	//userAvatarURLString := data.GetServerURL() + "/logo/external"
	//userAvatarURL, err := url.Parse(userAvatarURLString)
	//userAvatarURL.RawQuery = "uc=" + uniquenessString
	//if err != nil {
	//	log.Errorln("unable to parse user avatar url", userAvatarURLString, err)
	//}

	//image := streams.NewActivityStreamsImage()
	//imgProp := streams.NewActivityStreamsUrlProperty()
	//imgProp.AppendIRI(userAvatarURL)
	//image.SetActivityStreamsUrl(imgProp)
	//icon := streams.NewActivityStreamsIconProperty()
	//icon.AppendActivityStreamsImage(image)
	//app.SetActivityStreamsIcon(icon)

	// Actor  URL
	urlProperty := streams.NewActivityStreamsUrlProperty()
	urlProperty.AppendIRI(actorIRI)
	app.SetActivityStreamsUrl(urlProperty)

	// Profile header
	//headerImage := streams.NewActivityStreamsImage()
	//headerImgPropURL := streams.NewActivityStreamsUrlProperty()
	//headerImgPropURL.AppendIRI(userAvatarURL)
	//headerImage.SetActivityStreamsUrl(headerImgPropURL)
	//headerImageProp := streams.NewActivityStreamsImageProperty()
	//headerImageProp.AppendActivityStreamsImage(headerImage)
	//app.SetActivityStreamsImage(headerImageProp)

	// Profile bio
	//summaryProperty := streams.NewActivityStreamsSummaryProperty()
	//summaryProperty.AppendXMLSchemaString(config.GetServerSummary())
	//app.SetActivityStreamsSummary(summaryProperty)

	// Links
	//if serverURL := data.GetServerURL(); serverURL != "" {
	//	addMetadataLinkToProfile(app, "Stream", serverURL)
	//}
	//for _, link := range data.GetSocialHandles() {
	//	addMetadataLinkToProfile(app, link.Platform, link.URL)
	//}

	// Discoverable
	discoverableProperty := streams.NewTootDiscoverableProperty()
	discoverableProperty.Set(true)
	app.SetTootDiscoverable(discoverableProperty)

	// Followers
	followersProperty := streams.NewActivityStreamsFollowersProperty()
	followersURL := *actorIRI
	followersURL.Path = actorIRI.Path + "/followers"
	followersProperty.SetIRI(&followersURL)
	app.SetActivityStreamsFollowers(followersProperty)

	// Tags
	tagProp := streams.NewActivityStreamsTagProperty()
	//for _, tagString := range data.GetServerMetadataTags() {
	//	hashtag := MakeHashtag(tagString)
	//	tagProp.AppendTootHashtag(hashtag)
	//}

	app.SetActivityStreamsTag(tagProp)

	// Work around an issue where a single attachment will not serialize
	// as an array, so add another item to the mix.
	//if len(data.GetSocialHandles()) == 1 {
	//	addMetadataLinkToProfile(app, "Owncast", "https://owncast.online")
	//}

	return app
}
