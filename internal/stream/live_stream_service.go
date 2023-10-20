package stream

import (
	"context"
	"errors"
	"fmt"
	"path"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/auth"
	"github.com/shigde/sfu/internal/lobby"
)

var errStreamNotPartOfSpace = errors.New("stream is not part of space")
var errUserNotOwnerOfStream = errors.New("user is not owner of stream")

type LiveStreamService struct {
	streamRepo *LiveStreamRepository
	spaceRepo  *SpaceRepository
}

func NewLiveStreamService(repo *LiveStreamRepository, spaceRepo *SpaceRepository) *LiveStreamService {
	return &LiveStreamService{streamRepo: repo, spaceRepo: spaceRepo}
}

func (ls *LiveStreamService) CreateStreamAccessByVideo(ctx context.Context, video *models.Video) error {
	userId := buildFederatedId(video.Owner.PreferredUsername, video.Owner.GetActorIri().Host)
	channelId := buildFederatedId(video.Channel.PreferredUsername, video.Channel.GetActorIri().Host)

	streamID, _ := uuid.Parse(video.Uuid)

	// video.Guests
	ls.streamRepo.BuildGuestAccounts(ctx, video.Guests)

	account := &auth.Account{}
	account.Actor = video.Owner
	account.ActorId = video.Owner.ID
	account.User = userId
	account.UUID = uuid.NewString()

	lobbyEntity := lobby.NewLobbyEntity(streamID)

	space := &Space{}
	space.Account = account
	space.Channel = video.Channel
	space.Identifier = channelId

	stream := &LiveStream{}
	stream.Lobby = lobbyEntity
	stream.Account = account
	stream.Space = space
	stream.UUID = streamID
	stream.Video = video
	stream.User = userId

	if err := ls.streamRepo.UpsertLiveStream(ctx, stream); err != nil {
		return fmt.Errorf("upsert live stream: %w", err)
	}
	return nil
}

func (ls *LiveStreamService) UpdateStreamAccessByVideo(ctx context.Context, video *models.Video) error {
	if !ls.streamRepo.Contains(ctx, video.Uuid) {
		return ls.CreateStreamAccessByVideo(ctx, video)
	}
	// redundant but needed in case of update guests
	ls.streamRepo.BuildGuestAccounts(ctx, video.Guests)
	return nil
}

func (ls *LiveStreamService) DeleteStreamAccessByVideo(ctx context.Context, iri string) error {
	uuidString := path.Base(iri)
	videoUuid, err := uuid.Parse(uuidString)
	if err != nil {
		return fmt.Errorf("parsing video uuid: %w", err)
	}

	if err := ls.streamRepo.DeleteByUuid(ctx, videoUuid.String()); err != nil {
		return fmt.Errorf("deleting stream by uuid: %w", err)
	}

	return nil
}

func (ls *LiveStreamService) AllBySpaceIdentifier(ctx context.Context, identifier string) ([]LiveStream, error) {
	return ls.streamRepo.AllBySpaceIdentifier(ctx, identifier)
}

func (ls *LiveStreamService) FindByUuidAndSpaceIdentifier(ctx context.Context, uuid string, identifier string) (*LiveStream, error) {
	stream, err := ls.streamRepo.FindByUuid(ctx, uuid)
	if err != nil {
		return nil, fmt.Errorf("find stream by uuid: %w", err)
	}

	if stream.Space.Identifier != identifier {
		return nil, errStreamNotPartOfSpace
	}

	return stream, nil
}

func (ls *LiveStreamService) Delete(ctx context.Context, uuid string, user string) error {
	_, err := ls.getStreamByUser(ctx, uuid, user)
	if err != nil {
		return fmt.Errorf("find stream to delete by uuid: %w", err)
	}
	return ls.streamRepo.Delete(ctx, uuid)
}

func (ls *LiveStreamService) CreateStream(ctx context.Context, liveStream *LiveStream, identifier string, user string) (string, error) {
	space, err := ls.spaceRepo.GetSpaceByIdentifier(ctx, identifier)
	if err != nil {
		return "", fmt.Errorf("find space by identifier: %w", err)
	}
	if space.Account.User != user {
		return "", errUserNotOwnerOfStream
	}
	return ls.streamRepo.Add(ctx, liveStream)
}

func (ls *LiveStreamService) UpdateStream(ctx context.Context, liveStream *LiveStream, identifier string, user string) error {
	space, err := ls.spaceRepo.GetSpaceByIdentifier(ctx, identifier)
	if err != nil {
		return fmt.Errorf("find space by identifier: %w", err)
	}
	if space.Account.User != user {
		return errUserNotOwnerOfStream
	}
	return ls.streamRepo.Update(ctx, liveStream)
}

func (ls *LiveStreamService) getStreamByUser(ctx context.Context, uuid string, user string) (*LiveStream, error) {
	stream, err := ls.streamRepo.FindByUuid(ctx, uuid)
	if err != nil {
		return nil, fmt.Errorf("find stream by uuid: %w", err)
	}
	if stream.Account.User != user {
		return nil, errUserNotOwnerOfStream
	}
	return stream, nil
}
