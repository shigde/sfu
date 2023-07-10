package auth

import (
	"fmt"

	"github.com/google/uuid"
)

type Principal struct {
	UID       string `json:"uid"`
	SID       string `json:"sid"`
	Publish   bool   `json:"publish"`
	Subscribe bool   `json:"subscribe"`
}

func (principal *Principal) GetUid() string {
	return principal.UID
}

func (principal *Principal) GetUUid() (uuid.UUID, error) {
	var id uuid.UUID
	id, err := uuid.Parse(principal.UID)
	if err != nil {
		return id, fmt.Errorf("converting user uid to uuid %w", err)
	}
	return id, nil

}

func (principal *Principal) GetSid() string {
	return principal.SID
}

func (principal *Principal) IsPublisher() bool {
	return principal.Publish
}

func (principal *Principal) IsSubscriber() bool {
	return principal.Subscribe
}
