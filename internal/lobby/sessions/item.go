package sessions

import "github.com/google/uuid"

type Item struct {
	UserId uuid.UUID
	Done   chan bool
}

func NewItem(userId uuid.UUID) Item {
	return Item{
		UserId: userId,
		Done:   make(chan bool),
	}
}
