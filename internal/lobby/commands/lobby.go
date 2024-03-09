package commands

import "github.com/shigde/sfu/internal/lobby/sessions"

type LobbyCommand interface {
	execute(session *sessions.Session)
}
