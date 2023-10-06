package parser

import (
	"errors"

	"github.com/google/uuid"
)

type VideoProperties struct {
	// OriginallyPublishedAt *time.Time
	ShigActive      bool
	LatencyMode     uint
	Uuid            string
	State           uint
	Shig            *ShigGuests
	IsLiveBroadcast bool
	PermanentLive   bool
	LiveSaveReplay  bool
}

type ShigGuests struct {
	FirstGuest  string
	SecondGuest string
	ThirdGuest  string
}

func ExtractVideoUnknownProperties(withUnknown WithUnknown) (*VideoProperties, error) {
	unknownProps := withUnknown.GetUnknownProperties()

	video := &VideoProperties{}
	//-------
	//originallyPublishedAt, ok := unknownProps["originallyPublishedAt"].(*time.Time)
	//if !ok {
	//	return nil, errors.New("cannot parsing 'originallyPublishedAt'")
	//}
	//videoProps.OriginallyPublishedAt = originallyPublishedAt

	//-------
	latencyMode, ok := unknownProps["latencyMode"].(float64)
	if !ok {
		return nil, errors.New("cannot parsing 'latencyMode'")
	}
	video.LatencyMode = uint(latencyMode)

	//-------
	uuidString, ok := unknownProps["uuid"].(string)
	if !ok {
		return nil, errors.New("cannot parsing 'uuid'")
	}
	uuidObj, err := uuid.Parse(uuidString)
	if err != nil {
		return nil, errors.New("cannot converting  'uuid'")
	}
	video.Uuid = uuidObj.String()

	//-------
	state, ok := unknownProps["state"].(float64)
	if !ok {
		return nil, errors.New("cannot parsing 'state'")
	}
	video.State = uint(state)

	//-------
	isLiveBroadcast, ok := unknownProps["isLiveBroadcast"].(bool)
	if !ok {
		return nil, errors.New("cannot parsing 'isLiveBroadcast'")
	}
	video.IsLiveBroadcast = isLiveBroadcast

	//-------
	permanentLive, ok := unknownProps["permanentLive"].(bool)
	if !ok {
		return nil, errors.New("cannot parsing 'permanentLive'")
	}
	video.PermanentLive = permanentLive

	//-------
	liveSaveReplay, ok := unknownProps["liveSaveReplay"].(bool)
	if !ok {
		return nil, errors.New("cannot parsing 'liveSaveReplay'")
	}
	video.LiveSaveReplay = liveSaveReplay

	//-------
	shigData, ok := unknownProps["peertubeShig"].(map[string]interface{})
	if !ok {
		video.ShigActive = false
		return video, nil
	}

	return extractShigData(shigData, video)
}

func extractShigData(props map[string]interface{}, video *VideoProperties) (*VideoProperties, error) {
	video.Shig = &ShigGuests{}
	shigActive, ok := props["shigActive"].(bool)
	if !ok {
		return nil, errors.New("cannot parsing 'shig.shigActive'")
	}
	video.ShigActive = shigActive

	firstGuest, ok := props["firstGuest"].(string)
	if !ok {
		return nil, errors.New("cannot parsing 'shig.firstGuest'")
	}
	video.Shig.FirstGuest = firstGuest

	secondGuest, ok := props["secondGuest"].(string)
	if !ok {
		return nil, errors.New("cannot parsing 'shig.firstGuest'")
	}
	video.Shig.SecondGuest = secondGuest

	thirdGuest, ok := props["thirdGuest"].(string)
	if !ok {
		return nil, errors.New("cannot parsing 'shig.firstGuest'")
	}
	video.Shig.ThirdGuest = thirdGuest

	return video, nil
}

type WithUnknown interface {
	GetUnknownProperties() map[string]interface{}
}
