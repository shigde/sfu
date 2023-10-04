package remote

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

func FetchAccountAsActor(ctx context.Context, req *http.Request) (*models.Actor, error) {
	// Do not support redirects.
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("doing request: %w", err)
	}
	b, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading boosy request: %w", err)
	}

	var rawActivity map[string]interface{}
	if err := json.Unmarshal(b, &rawActivity); err != nil {
		return nil, fmt.Errorf("unmarshalling request body: %w", err)
	}

	t, err := streams.ToType(ctx, rawActivity)
	if err != nil {
		return nil, fmt.Errorf("determinding type: %w", err)
	}
	accountable, ok := t.(ActivityStreamAccount)
	if !ok {
		return nil, fmt.Errorf("casting type %T but is not a pub.Activity", t)
	}

	if accountable.GetJSONLDId() == nil {
		return nil, fmt.Errorf("incoming Activity %s did not have required id property set", accountable.GetTypeName())
	}

	actor, err := createActor(accountable)
	if err != nil {
		return nil, fmt.Errorf("casting to account activity pub activity")
	}

	return actor, nil
}

func createActor(accountable ActivityStreamAccount) (*models.Actor, error) {
	// first check if we actually already know this account
	uriProp := accountable.GetJSONLDId()
	if uriProp == nil || !uriProp.IsIRI() {
		return nil, errors.New("no id property found on person, or id was not an iri")
	}
	uri := uriProp.GetIRI()

	// we don't know the account, or we're being told to update it, so we need to generate it from the person -- at least we already have the URI!
	acct := &models.Actor{}
	acct.ActorIri = uri.String()

	// Username aka preferredUsername
	// We need this one so bail if it's not set.
	username, err := extractPreferredUsername(accountable)

	if err != nil {
		return nil, fmt.Errorf("couldn't extract username: %s", err)
	}
	acct.PreferredUsername = username

	acct.ActorType = accountable.GetTypeName()

	// InboxURI
	if accountable.GetActivityStreamsInbox() != nil && accountable.GetActivityStreamsInbox().GetIRI() != nil {
		acct.InboxIri = accountable.GetActivityStreamsInbox().GetIRI().String()
	}

	// SharedInboxURI:
	// only trust shared inbox if it has at least two domains,
	// from the right, in common with the domain of the account
	if sharedInboxURI := extractSharedInbox(accountable); sharedInboxURI != nil {
		acct.SharedInboxIri = sharedInboxURI.String()
	}

	// OutboxURI
	if accountable.GetActivityStreamsOutbox() != nil && accountable.GetActivityStreamsOutbox().GetIRI() != nil {
		acct.OutboxIri = accountable.GetActivityStreamsOutbox().GetIRI().String()
	}

	// FollowingURI
	if accountable.GetActivityStreamsFollowing() != nil && accountable.GetActivityStreamsFollowing().GetIRI() != nil {
		acct.FollowingIri = accountable.GetActivityStreamsFollowing().GetIRI().String()
	}

	// FollowersURI
	if accountable.GetActivityStreamsFollowers() != nil && accountable.GetActivityStreamsFollowers().GetIRI() != nil {
		acct.FollowersIri = accountable.GetActivityStreamsFollowers().GetIRI().String()
	}

	// publicKey
	pkey, pkeyURL, pkeyOwnerID, err := extractPublicKey(accountable)
	if err != nil {
		return nil, fmt.Errorf("couldn't get public key for person %s: %s", uri.String(), err)
	}

	if pkeyOwnerID.String() != acct.ActorIri {
		return nil, fmt.Errorf("public key %s was owned by %s and not by %s", pkeyURL, pkeyOwnerID, acct.ActorIri)
	}

	if acct.PublicKey, err = exportRsaPublicKeyAsPemStr(pkey); err != nil {
		return nil, fmt.Errorf("converting puplic key as string get wrong")
	}

	return acct, nil
}

func extractPreferredUsername(i withPreferredUsername) (string, error) {
	u := i.GetActivityStreamsPreferredUsername()
	if u == nil || !u.IsXMLSchemaString() {
		return "", errors.New("preferredUsername nil or not a string")
	}

	if u.GetXMLSchemaString() == "" {
		return "", errors.New("preferredUsername was empty")
	}

	return u.GetXMLSchemaString(), nil
}

func extractSharedInbox(withEndpoints withEndpoints) *url.URL {
	endpointsProp := withEndpoints.GetActivityStreamsEndpoints()
	if endpointsProp == nil {
		return nil
	}

	for iter := endpointsProp.Begin(); iter != endpointsProp.End(); iter = iter.Next() {
		if !iter.IsActivityStreamsEndpoints() {
			continue
		}

		endpoints := iter.Get()
		if endpoints == nil {
			continue
		}

		sharedInboxProp := endpoints.GetActivityStreamsSharedInbox()
		if sharedInboxProp == nil || !sharedInboxProp.IsIRI() {
			continue
		}

		return sharedInboxProp.GetIRI()
	}

	return nil
}

func extractPublicKey(i withPublicKey) (
	*rsa.PublicKey, // pubkey
	*url.URL, // pubkey ID
	*url.URL, // pubkey owner
	error,
) {
	pubKeyProp := i.GetW3IDSecurityV1PublicKey()
	if pubKeyProp == nil {
		return nil, nil, nil, errors.New("public key property was nil")
	}

	for iter := pubKeyProp.Begin(); iter != pubKeyProp.End(); iter = iter.Next() {
		if !iter.IsW3IDSecurityV1PublicKey() {
			continue
		}

		pkey := iter.Get()
		if pkey == nil {
			continue
		}

		pubKeyID, err := pub.GetId(pkey)
		if err != nil {
			continue
		}

		pubKeyOwnerProp := pkey.GetW3IDSecurityV1Owner()
		if pubKeyOwnerProp == nil {
			continue
		}

		pubKeyOwner := pubKeyOwnerProp.GetIRI()
		if pubKeyOwner == nil {
			continue
		}

		pubKeyPemProp := pkey.GetW3IDSecurityV1PublicKeyPem()
		if pubKeyPemProp == nil {
			continue
		}

		pkeyPem := pubKeyPemProp.Get()
		if pkeyPem == "" {
			continue
		}

		block, _ := pem.Decode([]byte(pkeyPem))
		if block == nil {
			continue
		}

		var p crypto.PublicKey
		switch block.Type {
		case "PUBLIC KEY":
			p, err = x509.ParsePKIXPublicKey(block.Bytes)
		case "RSA PUBLIC KEY":
			p, err = x509.ParsePKCS1PublicKey(block.Bytes)
		default:
			err = fmt.Errorf("unknown block type: %q", block.Type)
		}
		if err != nil {
			return nil, nil, nil, fmt.Errorf("parsing public key from block bytes: %w", err)
		}

		if p == nil {
			return nil, nil, nil, errors.New("returned public key was empty")
		}

		pubKey, ok := p.(*rsa.PublicKey)
		if !ok {
			continue
		}

		return pubKey, pubKeyID, pubKeyOwner, nil
	}

	return nil, nil, nil, errors.New("couldn't find public key")
}

func exportRsaPublicKeyAsPemStr(pubkey *rsa.PublicKey) (string, error) {
	pubkey_bytes, err := x509.MarshalPKIXPublicKey(pubkey)
	if err != nil {
		return "", err
	}
	pubkey_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pubkey_bytes,
		},
	)

	return string(pubkey_pem), nil
}

type ActivityStreamAccount interface {
	GetJSONLDId() vocab.JSONLDIdProperty
	GetTypeName() string
	GetActivityStreamsPreferredUsername() vocab.ActivityStreamsPreferredUsernameProperty
	GetActivityStreamsIcon() vocab.ActivityStreamsIconProperty
	GetActivityStreamsName() vocab.ActivityStreamsNameProperty
	GetActivityStreamsImage() vocab.ActivityStreamsImageProperty
	GetActivityStreamsSummary() vocab.ActivityStreamsSummaryProperty
	GetActivityStreamsAttachment() vocab.ActivityStreamsAttachmentProperty
	SetActivityStreamsSummary(vocab.ActivityStreamsSummaryProperty)
	GetTootDiscoverable() vocab.TootDiscoverableProperty
	GetActivityStreamsUrl() vocab.ActivityStreamsUrlProperty
	GetW3IDSecurityV1PublicKey() vocab.W3IDSecurityV1PublicKeyProperty
	GetActivityStreamsInbox() vocab.ActivityStreamsInboxProperty
	GetActivityStreamsOutbox() vocab.ActivityStreamsOutboxProperty
	GetActivityStreamsFollowing() vocab.ActivityStreamsFollowingProperty
	GetActivityStreamsFollowers() vocab.ActivityStreamsFollowersProperty
	GetTootFeatured() vocab.TootFeaturedProperty
	GetActivityStreamsManuallyApprovesFollowers() vocab.ActivityStreamsManuallyApprovesFollowersProperty
	GetActivityStreamsEndpoints() vocab.ActivityStreamsEndpointsProperty
	GetActivityStreamsTag() vocab.ActivityStreamsTagProperty
}

type withPreferredUsername interface {
	GetActivityStreamsPreferredUsername() vocab.ActivityStreamsPreferredUsernameProperty
}

type withEndpoints interface {
	GetActivityStreamsEndpoints() vocab.ActivityStreamsEndpointsProperty
}

type withPublicKey interface {
	GetW3IDSecurityV1PublicKey() vocab.W3IDSecurityV1PublicKeyProperty
}
