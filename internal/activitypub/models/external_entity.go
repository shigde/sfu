package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/superseriousbusiness/activity/streams/vocab"
)

type ExternalEntity interface {
	GetJSONLDId() vocab.JSONLDIdProperty
	GetActivityStreamsInbox() vocab.ActivityStreamsInboxProperty
	GetActivityStreamsName() vocab.ActivityStreamsNameProperty
	GetActivityStreamsPreferredUsername() vocab.ActivityStreamsPreferredUsernameProperty
	GetActivityStreamsIcon() vocab.ActivityStreamsIconProperty
	GetW3IDSecurityV1PublicKey() vocab.W3IDSecurityV1PublicKeyProperty
}

// MakeActorFromExernalAPEntity takes a full ActivityPub entity and returns our
// internal representation of an actor.
func MakeActorFromExernalAPEntity(entity ExternalEntity) (*Actor, error) {
	// Username is required (but not a part of the official ActivityPub spec)
	if entity.GetActivityStreamsPreferredUsername() == nil || entity.GetActivityStreamsPreferredUsername().GetXMLSchemaString() == "" {
		return nil, errors.New("remote activitypub entity does not have a preferred username set, rejecting")
	}
	// username := GetFullUsernameFromExternalEntity(entity)

	// Key is required
	if entity.GetW3IDSecurityV1PublicKey() == nil {
		return nil, errors.New("remote activitypub entity does not have a public key set, rejecting")
	}

	// Name is optional
	//var name string
	//if entity.GetActivityStreamsName() != nil && !entity.GetActivityStreamsName().Empty() {
	//	name = entity.GetActivityStreamsName().At(0).GetXMLSchemaString()
	//}

	apActor := Actor{
		ActorType:         "Person",
		PublicKey:         entity.GetW3IDSecurityV1PublicKey().Name(),
		PrivateKey:        sql.NullString{},
		ActorIri:          entity.GetJSONLDId().Get().String(),
		FollowingIri:      "",
		FollowersIri:      "",
		InboxIri:          entity.GetActivityStreamsInbox().GetIRI().String(),
		OutboxIri:         "",
		SharedInboxIri:    "",
		DisabledAt:        sql.NullTime{},
		ServerId:          sql.NullInt64{},
		RemoteCreatedAt:   time.Time{},
		PreferredUsername: entity.GetActivityStreamsPreferredUsername().GetXMLSchemaString(),
	}

	return &apActor, nil
}

func GetFullUsernameFromExternalEntity(entity ExternalEntity) string {
	hostname := entity.GetJSONLDId().GetIRI().Hostname()
	username := entity.GetActivityStreamsPreferredUsername().GetXMLSchemaString()
	fullUsername := fmt.Sprintf("%s@%s", username, hostname)

	return fullUsername
}
