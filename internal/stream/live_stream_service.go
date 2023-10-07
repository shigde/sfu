package stream

import (
	"context"
	"fmt"

	"github.com/shigde/sfu/internal/activitypub/models"
)

type LiveStreamService struct {
	repo *LiveStreamRepository
}

func NewLiveStreamService(repo *LiveStreamRepository) *LiveStreamService {
	return &LiveStreamService{repo: repo}
}

func (ls *LiveStreamService) CreateStreamAccessByVideo(ctx context.Context, video *models.Video) error {
	stream := &LiveStream{}
	// stream.UUID = video.Uuid

	if err := ls.repo.UpsertLiveStream(ctx, stream); err != nil {
		return fmt.Errorf("upsert live stream")
	}
	return nil
}

func (ls *LiveStreamService) DeleteStreamAccessByVideo(ctx context.Context, video *models.Video) error {
	return nil
}
