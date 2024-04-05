package sessions

import "github.com/google/uuid"

type Item struct {
	UserId      uuid.UUID
	SessionType SessionType
	Done        chan bool
}

func NewItem(userId uuid.UUID) Item {
	return Item{
		UserId:      userId,
		SessionType: UserSession,
		Done:        make(chan bool),
	}
}
