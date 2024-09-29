package migration

import (
	"fmt"
	"net/url"

	"github.com/shigde/sfu/internal/activitypub/instance"
	"github.com/shigde/sfu/internal/activitypub/models"
	"github.com/shigde/sfu/internal/auth"
	"github.com/shigde/sfu/internal/storage"
)

func LoadFixtures(config *instance.FederationConfig, storage storage.Storage) error {
	storage.GetDatabase()
	db := storage.GetDatabase()
	streamInstanceUrl, _ := url.Parse("https://stream.localhost:8080")
	remoteInstanceUrl, _ := url.Parse("https://remote.localhost:8070")

	streamInstance, err := buildInstance(db, streamInstanceUrl, "shig")
	if err != nil {
		return fmt.Errorf("fixtures build shig stream instance: %w", err)
	}

	// stream domain ---------------------------------------------------------------------------------------------------
	//  +-----------+   +-----------+                                               +-----------+   +-----------+
	//  |  <Actor>  |   |  <Actor>  |                                               |  <Actor>  |   |  <Actor>  |
	//  |  user123  |   |  user123  |                                               |   root    |   |   root    |
	//  |           |   |  channel  |                                               |           |   |  channel  |
	//  +-----+-----+   +-----+-----+                                               +-----+-----+   +-----+-----+
	//        |               |-------------+---------------+               +-------------|               |
	//        |               |             |               |               |             |               |
	//        +---------------^----------+--^-------------+ |               |  +----------^---------------+
	//        |               |          |  |             | |               |  |          |               |
	//        |               |       +--+--+-----+   +---+-+-----+   +-----+--+--+       |               |
	//        |               |       |  <Video>  |   |  <Video>  |   |  <Video>  |       |               |
	//        |               |       | fc9d575d  |   | ce33f40c  |   | 7b762908  |       |               |
	//        |               |       +-----------+   +-----------+   +-----------+       |               |
	//        |               |             |               |              |              |               |
	//        |               |             |               |              |              |               |
	//        V               V             |               |              |              V               V
	//  +-----------+   +-----------+       |               |              |          +-----------+   +-----------+
	//  | <Account> |   |  <Space>  |       |               |              |          | <Account> |   |  <Space>  |
	//  |  user123  + - +  user123  |       |               |              |          |   root    + - |   root    |
	//  |           |   |  channel  |       |               |              |          |           |   |  channel  |
	//  +-----+-----+   +------+----+       |               |              |          +-----+-----+   +-----+-----+
	//        |                |            |               |              |                |               |
	//        +----------------^--------+---^------------+  |              |      +---------+               |
	//                         |        |   |            |  |              |      |                         |
	//                         |        |   V            |  V              V      |                         |
	//                         |    +---+----------+   +-+-------------+   +------+------+                  |
	//                         |    | <LiveStream> |   | <LiveStream> |   | <LiveStream> |                  |
	//                         |    |   fc9d575d   |   |   ce33f40c   |   |   7b762908   |                  |
	//                         |    +------+-------+   +-------+------+   +------+-------+                  |
	//                         |           |                   |                 |                          |
	//                         --------+---^-------------+     |                 |   +----------------------+
	//                                 |   |             |     |                 |   |
	//                              +--+---+-------+   +-+-----+------+   +------+---+---+
	//                              |    <Lobby>   |   |    <Lobby>   |   |    <Lobby>   |
	//                              |   fc9d575d   |   |   ce33f40c   |   |   7b762908   |
	//                              +--------------+   +--------------+   +--------------+
	//
	// -----------------------------------------------------------------------------------------------------------------

	user123Actor, _ := models.NewPersonActor(streamInstanceUrl, "user123")
	rootActor, _ := models.NewPersonActor(streamInstanceUrl, "root")
	user123ChannelActor, _ := models.NewChannelActor(streamInstanceUrl, "user123_channel")
	rootChannelActor, _ := models.NewChannelActor(streamInstanceUrl, "root_channel")
	db.Save(user123Actor)
	db.Save(rootActor)
	db.Save(user123ChannelActor)
	db.Save(rootChannelActor)

	user123Account := auth.CreateAccount(creatUserId("user123", streamInstanceUrl), user123Actor, "96efea69-a084-4a33-9936-78d30c6301e8")
	rootAccount := auth.CreateAccount(creatUserId("root", streamInstanceUrl), rootActor, "ecd7c26a-f4ec-458b-8496-1a2834e50274")
	db.Save(user123Account)
	db.Save(rootAccount)

	video1 := NewVideo("live-stream-1", "fc9d575d-bc6f-46d6-9dc5-5b687889486f", streamInstance, user123Actor, user123ChannelActor)
	video2 := NewVideo("live-stream-2", "ce33f40c-0eeb-4b96-976a-26b5be0fa345", streamInstance, user123Actor, user123ChannelActor)
	video3 := NewVideo("live-stream-3", "7b762908-5a7f-49a6-9d05-ddcf26e8c07e", streamInstance, rootActor, rootChannelActor)

	liveStream1 := buildLiveStream(video1, user123Account)
	liveStream2 := buildLiveStream(video2, user123Account)
	liveStream3 := buildLiveStream(video3, rootAccount)
	db.Save(liveStream1)
	db.Save(liveStream2)
	db.Save(liveStream3)

	// remote domain ------------------------------------------------------------------------------
	//  +------------+   +------------+
	//  |  <Actor>   |   |  <Actor>   |
	//  | remoteUser |   | remoteUser |
	//  |            |   |  channel   |
	//  +-----+------+   +----+-------+
	//        |               |
	//        |               +-------------+
	//        |               |             |
	//        |---------------^----------+  |
	//        |               |          |  |
	//        |               |       +--+--+-----+
	//        |               |       |  <Video>  |
	//        |               |       | 034973c3  |
	//        |               |       +-----------+
	//        |               |             |
	//        V               V             |
	//  +------------+   +------------+     |
	//  | <Account>  |   |  <Space>   |     |
	//  | remoteUser | _ | remoteUser |     |
	//  |            |   |  channel   |     |
	//  +------------+   +------------+     |
	//        |                |            |
	//        +----------------^--------+   |
	//                         |        |   |
	//                         |        |   V
	//                         |    +---+----------+
	//                         |    | <LiveStream> |
	//                         |    |   034973c3   |
	//                         |    +------+-------+
	//                         |           |
	//                         +-------+   |
	//                                 |   |
	//                              +--+---+-------+
	//                              |    <Lobby>   |
	//                              |   034973c3   |
	//                              +--------------+
	//
	// -----------------------------------------------------------------------------------------------------------------

	remoteInstance, err := buildInstance(db, remoteInstanceUrl, "shig")
	if err != nil {
		return fmt.Errorf("fixtures build shig stream instance: %w", err)
	}

	remoteUserActor, _ := models.NewPersonActor(remoteInstanceUrl, "remoteUser")
	remoteUserChannelActor, _ := models.NewChannelActor(remoteInstanceUrl, "remoteUser_channel")
	db.Save(remoteUserActor)
	db.Save(remoteUserChannelActor)

	remoteUserAccount := auth.CreateAccount(creatUserId("remoteUser", remoteInstanceUrl), remoteUserActor, "517c225b-ae98-44fd-8ff6-2e2e4eeb7900")
	db.Save(remoteUserAccount)

	video4 := NewVideo("live-stream-4", "034973c3-1756-4de3-b565-96264aa893c2", remoteInstance, remoteUserActor, remoteUserChannelActor)

	liveStream4 := buildLiveStream(video4, remoteUserAccount)
	db.Save(liveStream4)

	return nil
}
