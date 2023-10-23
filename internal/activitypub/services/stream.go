package services

import (
	"context"

	"github.com/shigde/sfu/internal/activitypub/models"
)

type StreamService interface {
	CreateStreamAccessByVideo(ctx context.Context, video *models.Video) error
	UpdateStreamAccessByVideo(ctx context.Context, video *models.Video) error
	DeleteStreamAccessByVideo(ctx context.Context, iri string) error
}
