package migration

import (
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/auth"
	"github.com/shigde/sfu/internal/lobby"
	"github.com/shigde/sfu/internal/storage"
	"github.com/shigde/sfu/internal/stream"
)

func LoadFixtures(config *instance.FederationConfig, storage storage.Storage) error {
	storage.GetDatabase()
	db := storage.GetDatabase()
	streamInstanceUrl, _ := url.Parse("https://stream.localhost:8080")
	remoteInstanceUrl, _ := url.Parse("https://remote.localhost:8070")
	// instanceActorIri := instance.BuildAccountIri(config.InstanceUrl, config.InstanceUsername)

	streamInstance, err := buildInstance(db, streamInstanceUrl, "shig")
	if err != nil {
		return fmt.Errorf("fixtures build shig stream instance: %w", err)
	}

	//var instanceActor *models.Actor
	//result := db.Where("actor_iri=?", instanceActorIri.String()).First(savedActor)
	//if result.Error != nil {
	//	return fmt.Errorf("fixtures loading actor: %w", result.Error)
	//}
	//
	//var instance *models.Instance
	//result = db.Where("actor_id=?", instanceActor.ID).First(instance)
	//if result.Error != nil {
	//	return fmt.Errorf("fixtures loading actor: %w", result.Error)
	//}

	// stream domain ------------------------------------------------------------------------------
	//  +-----------+   +-----------+                                               +-----------+   +-----------+
	//  |  <Actor>  |   |  <Actor>  |                                               |  <Actor>  |   |  <Actor>  |
	//  |  user123  |   |  user123  |                                               |   root    |   |   root    |
	//  |           |   |  channel  |                                               |           |   |  channel  |
	//  +-----------+   +-----------+                                               +-----------+   +-----------+
	//        |               |______________________________                _____________|               |
	//        |               |             |               |               |             |               |
	//        |_______________^_____________^______________ |               |   __________^_______________|
	//        |               |          |  |             | |               |  |          |               |
	//        |               |       +-----------+   +-----------+   +-----------+       |               |
	//        |               |       |  <Video>  |   |  <Video>  |   |  <Video>  |       |               |
	//        |               |       | fc9d575d  |   | ce33f40c  |   | 7b762908  |       |               |
	//        |               |       +-----------+   +-----------+   +-----------+       |               |
	//        |               |                                                           |               |
	//        |               |                                                           |               |
	//  +-----------+   +-----------+                                                +-----------+   +-----------+
	//  | <Account> |   |  <Space>  |                                                | <Account> |   |  <Space>  |
	//  |  user123  | _ |  user123  |                                                |   root    | _ |   root    |
	//  |           |   |  channel  |                                                |           |   |  channel  |
	//  +-----------+   +-----------+                                                +-----------+   +-----------+
	//            __________⁄  |
	//
	//
	user123Actor, _ := models.NewPersonActor(streamInstanceUrl, "user123")
	rootActor, _ := models.NewPersonActor(streamInstanceUrl, "root")
	user123ChannelActor, _ := models.NewChannelActor(streamInstanceUrl, "user123_channel")
	rootChannelActor, _ := models.NewChannelActor(streamInstanceUrl, "root_channel")
	db.Save(user123Actor)
	db.Save(rootActor)
	db.Save(user123ChannelActor)
	db.Save(rootChannelActor)

	//if config.InstanceUrl.String() == streamInstanceUrl.String() {
	user123Account := auth.CreateAccount(creatUserId("user123", streamInstanceUrl), user123Actor, "96efea69-a084-4a33-9936-78d30c6301e8")
	rootAccount := auth.CreateAccount(creatUserId("root", streamInstanceUrl), rootActor, "ecd7c26a-f4ec-458b-8496-1a2834e50274")
	db.Save(user123Account)
	db.Save(rootAccount)

	video1 := NewVideo("live-stream-1", "fc9d575d-bc6f-46d6-9dc5-5b687889486f", streamInstance, user123Actor, user123ChannelActor)
	//video2 := NewVideo("live-stream-2", "ce33f40c-0eeb-4b96-976a-26b5be0fa345", streamInstance, user123Actor, user123ChannelActor)
	//video3 := NewVideo("live-stream-3", "7b762908-5a7f-49a6-9d05-ddcf26e8c07e", streamInstance, rootActor, rootChannelActor)

	streamID, _ := uuid.Parse(video1.Uuid)
	space := stream.NewSpace(video1.Channel, user123Account)
	lobbyEntity := lobby.NewLobbyEntity(streamID, space.Identifier, video1.Instance.Actor.ActorIri)
	liveStream := stream.NewLiveStream(user123Account, lobbyEntity, space, video1)
	db.Save(liveStream)
	//}

	// remote domain ------------------------------------------------------------------------------
	//  +------------+   +------------+
	//  |  <Actor>   |   |  <Actor>   |
	//  | remoteUser |   | remoteUser |
	//  |            |   |  channel   |
	//  +------------+   +------------+
	//        |               |______________
	//        |               |             |
	//        |_______________^___________  |
	//        |               |          |  |
	//        |               |       +-----------+
	//        |               |       |  <Video>  |
	//        |               |       | 034973c3  |
	//        |               |       +-----------+
	//        |               |
	//        |               |
	//  +------------+   +------------+
	//  | <Account>  |   |  <Space>   |
	//  | remoteUser | _ | remoteUser |
	//  |            |   |  channel   |
	//  +------------+   +------------+
	//
	remoteUserActor, _ := models.NewPersonActor(remoteInstanceUrl, "remoteUser")
	remoteUserChannelActor, _ := models.NewChannelActor(remoteInstanceUrl, "remoteUser_channel")
	db.Save(remoteUserActor)
	db.Save(remoteUserChannelActor)

	// video 4:  034973c3-1756-4de3-b565-96264aa893c2
	if config.InstanceUrl.String() == remoteInstanceUrl.String() {
		remoteUserAccount := auth.CreateAccount(creatUserId("remoteUser", remoteInstanceUrl), remoteUserActor, "517c225b-ae98-44fd-8ff6-2e2e4eeb7900")
		db.Save(remoteUserAccount)
	}

	return nil
}
