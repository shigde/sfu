package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/activitypub/parser"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

type VideoService struct {
	config       *instance.FederationConfig
	actorService *ActorService
	videoRep     *models.VideoRepository
}

func NewVideoService(config *instance.FederationConfig, actorService *ActorService, videoRep *models.VideoRepository) *VideoService {
	return &VideoService{config: config, actorService: actorService, videoRep: videoRep}
}

func (s *VideoService) UpsertVideo(ctx context.Context, updateObject vocab.ActivityStreamsObjectProperty) error {

	for iter := updateObject.Begin(); iter != updateObject.End(); iter = iter.Next() {
		if iter.GetType().GetTypeName() == "Video" {
			var video models.Video
			// HANDLE VIDEO
			asVideo, ok := iter.GetType().(vocab.ActivityStreamsVideo)
			if !ok {
				return errors.New("couldn't parse video into vocab.ActivityStreamsVideo")
			}

			// VideoID
			videoIri := asVideo.GetJSONLDId()
			if !videoIri.IsIRI() {
				return errors.New("vocab.ActivityStreamsVideo has no iri")
			}
			video.Iri = videoIri.Get().String()

			owners, err := parser.ExtractAttributedTo(asVideo)
			if err != nil {
				return fmt.Errorf("parsing owners of video: %w", err)
			}
			// Owner
			_ = len(owners)

			published := parser.ExtractPublished(asVideo)
			video.Published = sql.NullTime{Time: published}

			video.Name = parser.ExtractName(asVideo)

			videoProps, err := parser.ExtractVideoUnknownProperties(asVideo)
			if err != nil {
				return fmt.Errorf("parsing unknown properties of video: %w", err)
			}
			video.LatencyMode = videoProps.LatencyMode
			video.Uuid = videoProps.Uuid
			video.State = videoProps.State
			video.IsLiveBroadcast = videoProps.IsLiveBroadcast
			video.PermanentLive = videoProps.PermanentLive
			video.LiveSaveReplay = videoProps.LiveSaveReplay
			video.ShigActive = videoProps.ShigActive

			//if videoProps.ShigActive {
			//	// fetching guest actors
			//
			//	//if len(videoProps.Shig.FirstGuest) > 0 {
			//	//
			//	//}
			//	//if len(videoProps.Shig.SecondGuest) > 0 {
			//	//
			//	//}
			//	//if len(videoProps.Shig.ThirdGuest) > 0 {
			//	//
			//	//}
			//}
		}
	}
	return nil
}
