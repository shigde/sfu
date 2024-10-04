package session

import (
	"fmt"

	"github.com/google/uuid"
)

type Principal struct {
	UUID string `json:"uuid"`
}

func (principal *Principal) GetUuidString() string {
	return principal.UUID
}

func (principal *Principal) GetUuid() (uuid.UUID, error) {
	var id uuid.UUID
	id, err := uuid.Parse(principal.UUID)
	if err != nil {
		return id, fmt.Errorf("converting user uid to uuid %w", err)
	}
	return id, nil

}
