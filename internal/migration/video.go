package migration

import "github.com/shigde/sfu/internal/activitypub/models"

func NewVideo(name string, uuidStr string, instance *models.Instance, owner *models.Actor, channel *models.Actor) *models.Video {
	video := &models.Video{
		Iri:             "https://peertube.localhost:9001/videos/watch/" + uuidStr,
		Uuid:            uuidStr,
		Name:            name,
		ShigActive:      true,
		Instance:        instance,
		InstanceId:      instance.ID,
		Owner:           owner,
		OwnerId:         owner.ID,
		Channel:         channel,
		ChannelId:       channel.ID,
		Guests:          nil,
		IsLiveBroadcast: true,
		LiveSaveReplay:  true,
		PermanentLive:   true,
		LatencyMode:     1,
		State:           4,
	}
	return video
}
