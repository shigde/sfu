package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"

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

func (s *VideoService) AddVideo(ctx context.Context, updateObject vocab.ActivityStreamsObjectProperty, toFollowerIris []*url.URL) error {
	video := &models.Video{}

	owners := s.buildOwnerIrisFromFollower(toFollowerIris)
	instAct, err := s.actorService.GetLocalInstanceActor(ctx)
	if err != nil {
		return fmt.Errorf("getting local instance actor: %w", err)
	}
	if err := s.addOwnerAndChannel(ctx, owners, video, instAct); err != nil {
		return fmt.Errorf("determining owner of video: %w", err)
	}

	for iter := updateObject.Begin(); iter != updateObject.End(); iter = iter.Next() {
		if iter.IsIRI() {
			videoIriObjectIRI := iter.GetIRI()

			req, err := s.actorService.sender.GetSignedRequest(instAct.GetActorIri(), videoIriObjectIRI.String())
			if err != nil {
				return fmt.Errorf("getting signed request object: %w", err)
			}
			rawVideo, err := s.actorService.sender.DoRequest(req)
			if err != nil {
				return fmt.Errorf("requesting video object: %w", err)
			}

			video.Iri = videoIriObjectIRI.String()

			video.Published = parser.ExcludeUnknownNullTime(rawVideo, "published")

			if video.Name, err = parser.ExcludeUnknownString(rawVideo, "name"); err != nil {
				return fmt.Errorf("getting video name: %w", err)
			}

			if err := s.parseCommonUnknownProps(ctx, rawVideo, video, instAct); err != nil {
				return fmt.Errorf("building common ideo props owner of video: %w", err)
			}
		}
	}

	if _, err := s.videoRep.Upsert(ctx, video); err != nil {
		return fmt.Errorf("creating video: %w", err)
	}

	return nil
}

func (s *VideoService) UpsertVideo(ctx context.Context, updateObject vocab.ActivityStreamsObjectProperty) error {
	video := &models.Video{}
	for iter := updateObject.Begin(); iter != updateObject.End(); iter = iter.Next() {

		if iter.GetType() == nil {
			continue
		}

		if iter.GetType().GetTypeName() == "Video" {
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
			instAct, err := s.actorService.GetLocalInstanceActor(ctx)
			if err != nil {
				return fmt.Errorf("getting local instance actor: %w", err)
			}

			if err := s.addOwnerAndChannel(ctx, owners, video, instAct); err != nil {
				return fmt.Errorf("determining owner of video: %w", err)
			}

			published := parser.ExtractPublished(asVideo)
			video.Published = sql.NullTime{Time: published}

			video.Name = parser.ExtractName(asVideo)

			if err := s.parseCommonUnknownProps(ctx, asVideo.GetUnknownProperties(), video, instAct); err != nil {
				return fmt.Errorf("building common ideo props owner of video: %w", err)
			}
		}
	}

	if _, err := s.videoRep.Upsert(ctx, video); err != nil {
		return fmt.Errorf("saving video: %w", err)
	}

	return nil
}

func (s *VideoService) addOwnerAndChannel(ctx context.Context, owners []*url.URL, video *models.Video, instAct *models.Actor) error {

	for _, iri := range owners {
		if actor, err := s.createActor(ctx, iri, instAct); err == nil {
			if actor.GetActorType() == models.Person {
				if video.Owner != nil {
					return fmt.Errorf("more then one owners dbId_1 %d, dbId_2 %d", video.Owner.ID, actor.ID)
				}
				video.Owner = actor
			}
			if actor.GetActorType() == models.Group {
				if video.Channel != nil {
					return fmt.Errorf("more then one channels dbId_1 %d, dbId_2 %d", video.Owner.ID, actor.ID)
				}
				video.Channel = actor
			}
		}
	}
	return nil
}

func (s *VideoService) addGuest(ctx context.Context, guest string, video *models.Video, instAct *models.Actor) {
	if len(guest) > 0 {
		if accountIri, err := s.buildAccountIri(guest); err == nil {
			if guest, err := s.createActor(ctx, accountIri, instAct); err == nil {
				video.Guests = append(video.Guests, guest)
			}
		}
	}
}

func (s *VideoService) createActor(ctx context.Context, accountIri *url.URL, instAct *models.Actor) (*models.Actor, error) {
	return s.actorService.CreateActorFromRemoteAccount(ctx, accountIri.String(), instAct)
}

func (s *VideoService) buildAccountIri(accountName string) (*url.URL, error) {
	names := strings.Split(accountName, "@")
	if len(names) != 2 {
		return nil, errors.New("no valid account name")
	}
	iri, err := url.Parse("/accounts/" + names[0])
	if err != nil {
		return nil, errors.New("no valid account name")
	}
	iri.Host = names[1]
	iri.Scheme = "https"
	if !s.config.Https {
		iri.Scheme = "http"
	}
	return iri, nil
}

func (s *VideoService) parseCommonUnknownProps(ctx context.Context, asVideoProps map[string]interface{}, video *models.Video, instAct *models.Actor) error {
	videoProps, err := parser.ExtractVideoUnknownProperties(asVideoProps)
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

	if videoProps.ShigActive {
		s.addGuest(ctx, videoProps.Shig.FirstGuest, video, instAct)
		s.addGuest(ctx, videoProps.Shig.SecondGuest, video, instAct)
		s.addGuest(ctx, videoProps.Shig.ThirdGuest, video, instAct)
	}
	return nil
}

func (s *VideoService) buildOwnerIrisFromFollower(followerList []*url.URL) []*url.URL {
	owners := make([]*url.URL, 0)
	for _, follower := range followerList {
		if strings.HasSuffix(follower.Path, "followers") {
			follower.Path = strings.TrimSuffix(follower.Path, "followers")
			owners = append(owners, follower)
		}
	}

	return owners
}
