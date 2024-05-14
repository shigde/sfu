package auth

import (
	"crypto/md5"
	"fmt"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/activitypub/models"
)

func CreateInstanceAccount(userId string, actor *models.Actor) *Account {
	md5Uuid := CreateInstanceUuid(userId)
	return &Account{
		User: userId,
		UUID: md5Uuid.String(),
		// ActorId: actor.ID,
		Actor: actor,
	}
}

func CreateInstanceUuid(userId string) uuid.UUID {
	nameByte := []byte(userId)
	md5String := fmt.Sprintf("%x", md5.Sum(nameByte))
	return uuid.MustParse(md5String)
}

func CreateAccount(actorId string, actor *models.Actor, uuidStr string) *Account {
	return &Account{
		User:    actorId,
		UUID:    uuidStr,
		ActorId: actor.ID,
		Actor:   actor,
	}
}
