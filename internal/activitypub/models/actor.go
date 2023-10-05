package models

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/shigde/sfu/internal/activitypub/crypto"
	"github.com/shigde/sfu/internal/activitypub/instance"
	"gorm.io/gorm"
)

var (
	ErrorNoPreferredUsername = errors.New("remote activitypub entity does not have a preferred username set, rejecting")
	ErrorNoPublicKey         = errors.New("remote activitypub entity does not have a public key set, rejecting")
)

type Actor struct {
	ActorType         string         `gorm:"type"`
	PublicKey         string         `gorm:"publicKey"`
	PrivateKey        sql.NullString `gorm:"privateKey"`
	ActorIri          string         `gorm:"actorIri"`
	FollowingIri      string         `gorm:"followingIri"`
	FollowersIri      string         `gorm:"followersIri"`
	InboxIri          string         `gorm:"inboxIri"`
	OutboxIri         string         `gorm:"outboxIri"`
	SharedInboxIri    string         `gorm:"sharedInboxIri"`
	DisabledAt        sql.NullTime   `gorm:"disabledAt"`
	ServerId          sql.NullInt64  `gorm:"serverId"`
	RemoteCreatedAt   time.Time      `gorm:"remoteCreatedAt"`
	PreferredUsername string         `gorm:"preferredUsername"`
	Follower          []Follow       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Following         []Follow       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	gorm.Model
}

func newInstanceActor(instanceUrl *url.URL, name string) (*Actor, error) {
	actorIri := instance.BuildAccountIri(instanceUrl, name)
	now := time.Now()
	privateKey, publicKey, err := crypto.GenerateKeys()
	if err != nil {
		return nil, fmt.Errorf("generation key pair")
	}
	return &Actor{
		ActorType:         "Application",
		PublicKey:         string(publicKey),
		PrivateKey:        sql.NullString{String: string(privateKey), Valid: true},
		ActorIri:          actorIri.String(),
		FollowingIri:      instance.BuildFollowingIri(actorIri).String(),
		FollowersIri:      instance.BuildFollowersIri(actorIri).String(),
		InboxIri:          instance.BuildInboxIri(actorIri).String(),
		OutboxIri:         instance.BuildOutboxIri(actorIri).String(),
		SharedInboxIri:    instance.BuildSharedInboxIri(instanceUrl).String(),
		DisabledAt:        sql.NullTime{},
		RemoteCreatedAt:   now,
		PreferredUsername: name,
	}, nil
}

func (s *Actor) GetActorIri() *url.URL {
	iri, _ := url.Parse(s.ActorIri)
	return iri
}

func (s *Actor) GetInboxIri() *url.URL {
	iri, _ := url.Parse(s.InboxIri)
	return iri
}

func (s *Actor) GetOutboxIri() *url.URL {
	iri, _ := url.Parse(s.OutboxIri)
	return iri
}

type ActorType uint

const (
	Person ActorType = iota
	Group
	Organization
	Application
	Service
)

func (at ActorType) String() string {
	return []string{"Person", "Group", "Organization", "Application", "Service"}[at]
}

func ActorTypeFromString(str string) ActorType {
	switch strings.ToLower(str) {
	case "person":
		return Person
	case "group":
		return Group
	case "organization":
		return Organization
	case "application":
		return Application
	case "service":
		return Service
	}
	return Person
}
