package lobby

import "github.com/google/uuid"

type lobbyItem struct {
	LobbyId uuid.UUID
	Done    chan bool
}

func newLobbyItem(userId uuid.UUID) lobbyItem {
	return lobbyItem{
		LobbyId: userId,
		Done:    make(chan bool),
	}
}
