package migration

import (
	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/auth"
	"github.com/shigde/sfu/internal/lobby"
	"github.com/shigde/sfu/internal/stream"
)

func buildLiveStream(video *models.Video, account *auth.Account) *stream.LiveStream {
	streamID, _ := uuid.Parse(video.Uuid)
	space := stream.NewSpace(video.Channel, account)
	lobbyEntity := lobby.NewLobbyEntity(streamID, space.Identifier, video.Instance.Actor.ActorIri)
	return stream.NewLiveStream(account, lobbyEntity, space, video)
}
