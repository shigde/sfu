package auth

import (
	"crypto/md5"
	"fmt"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/activitypub/models"
)

func CreateInstanceAccount(actorId string, actor *models.Actor) *Account {
	md5Uuid := CreateShigInstanceId(actorId)
	return &Account{
		User:    actorId,
		UUID:    md5Uuid.String(),
		ActorId: actor.ID,
		Actor:   actor,
	}
}

func CreateShigInstanceId(_ string) uuid.UUID {
	actorId := "test-this-out"
	nameByte := []byte(actorId)
	md5String := fmt.Sprintf("%x", md5.Sum(nameByte))
	return uuid.MustParse(md5String)
}
