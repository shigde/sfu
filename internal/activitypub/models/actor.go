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
	ActorType         string         `gorm:""`
	PublicKey         string         `gorm:""`
	PrivateKey        sql.NullString `gorm:""`
	ActorIri          string         `gorm:"index;unique;"`
	FollowingIri      string         `gorm:""`
	FollowersIri      string         `gorm:""`
	InboxIri          string         `gorm:""`
	OutboxIri         string         `gorm:""`
	SharedInboxIri    string         `gorm:""`
	DisabledAt        sql.NullTime   `gorm:""`
	ServerId          sql.NullInt64  `gorm:""`
	RemoteCreatedAt   time.Time      `gorm:""`
	PreferredUsername string         `gorm:""`
	Follower          []*Follow      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Following         []*Follow      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	VideoGuest        []*Video       `gorm:"many2many:video_guests;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	gorm.Model
}

func NewInstanceActor(instanceUrl *url.URL, name string) (*Actor, error) {
	actorIri := instance.BuildAccountIri(instanceUrl, name)
	privateKey, publicKey, err := crypto.GenerateKeys()
	if err != nil {
		return nil, fmt.Errorf("generation key pair")
	}
	return newActor(Application, actorIri, instanceUrl, name, publicKey, privateKey), nil
}

func NewTrustedInstanceActor(actorIri *url.URL, name string) (*Actor, error) {
	instanceUrl, err := url.Parse(fmt.Sprintf("%s://%s", actorIri.Scheme, actorIri.Host))
	if err != nil {
		return nil, fmt.Errorf("trusted instance actor instanceUrl url")
	}
	return newActor(Application, actorIri, instanceUrl, name, make([]byte, 0), make([]byte, 0)), nil
}

func NewPersonActor(instanceUrl *url.URL, name string) (*Actor, error) {
	actorIri := instance.BuildAccountIri(instanceUrl, name)
	privateKey, publicKey, err := crypto.GenerateKeys()
	if err != nil {
		return nil, fmt.Errorf("generation key pair")
	}
	return newActor(Person, actorIri, instanceUrl, name, publicKey, privateKey), nil
}

func NewChannelActor(instanceUrl *url.URL, name string) (*Actor, error) {
	actorIri := instance.BuildAccountIri(instanceUrl, name)
	privateKey, publicKey, err := crypto.GenerateKeys()
	if err != nil {
		return nil, fmt.Errorf("generation key pair")
	}
	return newActor(Group, actorIri, instanceUrl, name, publicKey, privateKey), nil
}

func newActor(actorType ActorType, actorIri *url.URL, instanceUrl *url.URL, name string, publicKey []byte, privateKey []byte) *Actor {
	now := time.Now()

	var privateKeyStr = sql.NullString{}
	if len(privateKey) > 0 {
		privateKeyStr = sql.NullString{String: string(privateKey), Valid: true}
	}
	var publicKeyStr = ""
	if len(publicKey) > 0 {
		publicKeyStr = string(publicKey)
	}

	return &Actor{
		ActorType:         actorType.String(),
		PublicKey:         publicKeyStr,
		PrivateKey:        privateKeyStr,
		ActorIri:          actorIri.String(),
		FollowingIri:      instance.BuildFollowingIri(actorIri).String(),
		FollowersIri:      instance.BuildFollowersIri(actorIri).String(),
		InboxIri:          instance.BuildInboxIri(actorIri).String(),
		OutboxIri:         instance.BuildOutboxIri(actorIri).String(),
		SharedInboxIri:    instance.BuildSharedInboxIri(instanceUrl).String(),
		DisabledAt:        sql.NullTime{},
		RemoteCreatedAt:   now,
		PreferredUsername: name,
	}
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

func (s *Actor) GetSharedInboxIri() *url.URL {
	iri, _ := url.Parse(s.SharedInboxIri)
	return iri
}

func (s *Actor) GetActorType() ActorType {
	return ActorTypeFromString(s.ActorType)
}

type ActorType uint

const (
	Person ActorType = iota
	Group
	Organization
	Application
	Service
	Bot
)

func (at ActorType) String() string {
	return []string{"Person", "Group", "Organization", "Application", "Service", "Bot"}[at]
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
	return Bot
}
