package stream

import (
	"context"
	"fmt"
	"path"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/auth"
)

type LiveStreamService struct {
	repo *LiveStreamRepository
}

func NewLiveStreamService(repo *LiveStreamRepository) *LiveStreamService {
	return &LiveStreamService{repo: repo}
}

func (ls *LiveStreamService) CreateStreamAccessByVideo(ctx context.Context, video *models.Video) error {
	userId := fmt.Sprintf("%s@%s", video.Owner.PreferredUsername, video.Owner.GetActorIri().Host)
	channelId := fmt.Sprintf("%s@%s", video.Channel.PreferredUsername, video.Channel.GetActorIri().Host)

	account := &auth.Account{}
	account.Actor = video.Owner
	account.User = userId
	account.UUID = uuid.NewString()

	space := &Space{}
	space.Account = account
	space.Channel = video.Channel
	space.Identifier = channelId

	stream := &LiveStream{}
	stream.Account = account
	stream.Space = space
	stream.UUID, _ = uuid.Parse(video.Uuid)
	stream.Video = video
	stream.User = userId

	if err := ls.repo.UpsertLiveStream(ctx, stream); err != nil {
		return fmt.Errorf("upsert live stream: %w", err)
	}
	return nil
}

func (ls *LiveStreamService) DeleteStreamAccessByVideo(ctx context.Context, iri string) error {
	uuidString := path.Base(iri)
	videoUuid, err := uuid.Parse(uuidString)
	if err != nil {
		return fmt.Errorf("parsing video uuid: %w", err)
	}

	if err := ls.repo.DeleteByUuid(ctx, videoUuid.String()); err != nil {
		return fmt.Errorf("deleting stream by uuid: %w", err)
	}

	return nil
}
