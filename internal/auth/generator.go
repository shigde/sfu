package auth

import (
	"crypto/md5"
	"fmt"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/activitypub/models"
)

func CreateInstanceAccount(name string, actor *models.Actor) *Account {
	nameByte := []byte(name)
	md5String := fmt.Sprintf("%x", md5.Sum(nameByte))

	md5Uuid := uuid.MustParse(md5String)
	return &Account{
		User:    name,
		UUID:    md5Uuid.String(),
		ActorId: actor.ID,
		Actor:   actor,
	}
}
