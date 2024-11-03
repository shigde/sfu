package account

import (
	"crypto/md5"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/activitypub/models"
)

func CreateInstanceAccount(userId string, actor *models.Actor) *Account {
	md5Uuid := CreateInstanceUuid(userId)
	return &Account{
		User: userId,
		UUID: md5Uuid.String(),
		// ActorId: actor.ID,
		Actor: actor,
		Role:  SERVICE,
		Email: PlaceholderEmail(uuid.NewString()),
	}
}

func CreateInstanceUuid(userId string) uuid.UUID {
	nameByte := []byte(userId)
	md5String := fmt.Sprintf("%x", md5.Sum(nameByte))
	return uuid.MustParse(md5String)
}

func CreateAccount(email string, actor *models.Actor, uuidStr string) *Account {
	userId := instance.BuildUserId(actor.PreferredUsername, actor.GetActorIri())
	return &Account{
		User:    userId,
		Email:   email,
		UUID:    uuidStr,
		ActorId: actor.ID,
		Actor:   actor,
		Role:    USER,
	}
}

func CreateAdminAccount(email string, actor *models.Actor, uuidStr string) *Account {
	acc := CreateAccount(email, actor, uuidStr)
	acc.Role = ADMIN
	return acc
}

func CreateGuestAccount(email string, actor *models.Actor, uuidStr string) *Account {
	acc := CreateAccount(email, actor, uuidStr)
	acc.Role = GUEST
	return acc
}

func PlaceholderEmail(counter string) string {
	return fmt.Sprintf("%s-placeholder@shig.de", counter)
}

func IsPlaceholderEmail(mail string) bool {
	return strings.HasSuffix(mail, "placeholder@shig.de")
}
