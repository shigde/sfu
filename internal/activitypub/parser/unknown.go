package parser

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type VideoPropertiesUnknown struct {
	// OriginallyPublishedAt *time.Time
	LatencyMode     uint
	Uuid            string
	State           uint
	ShigActive      bool
	Shig            *ShigGuests
	ShigInstanceUrl string
	IsLiveBroadcast bool
	PermanentLive   bool
	LiveSaveReplay  bool
}

type ShigGuests struct {
	FirstGuest  string
	SecondGuest string
	ThirdGuest  string
}

func ExtractVideoUnknownProperties(unknownProps map[string]interface{}) (*VideoPropertiesUnknown, error) {
	video := &VideoPropertiesUnknown{}
	var err error

	//-------
	if video.LatencyMode, err = ExcludeUnknownUint(unknownProps, "latencyMode"); err != nil {
		return nil, err
	}

	uuidVal, err := ExcludeUnknownUuid(unknownProps, "uuid")
	if err != nil {
		return nil, err
	}
	video.Uuid = uuidVal.String()

	if video.State, err = ExcludeUnknownUint(unknownProps, "state"); err != nil {
		return nil, err
	}

	if video.IsLiveBroadcast, err = ExcludeUnknownBool(unknownProps, "isLiveBroadcast"); err != nil {
		return nil, err
	}

	if video.PermanentLive, err = ExcludeUnknownBool(unknownProps, "permanentLive"); err != nil {
		return nil, err
	}

	if video.LiveSaveReplay, err = ExcludeUnknownBool(unknownProps, "liveSaveReplay"); err != nil {
		return nil, err
	}

	//-------
	shigData, ok := unknownProps["peertubeShig"].(map[string]interface{})
	if !ok {
		video.ShigActive = false
		return video, nil
	}

	return extractShigData(shigData, video)
}

func extractShigData(props map[string]interface{}, video *VideoPropertiesUnknown) (*VideoPropertiesUnknown, error) {
	video.Shig = &ShigGuests{}

	var err error
	if video.ShigActive, err = ExcludeUnknownBool(props, "shigActive"); err != nil {
		return nil, err
	}

	if video.ShigInstanceUrl, err = ExcludeUnknownString(props, "shigInstanceUrl"); err != nil {
		return nil, err
	}

	if video.Shig.FirstGuest, err = ExcludeUnknownString(props, "firstGuest"); err != nil {
		return nil, err
	}

	if video.Shig.SecondGuest, err = ExcludeUnknownString(props, "secondGuest"); err != nil {
		return nil, err
	}

	if video.Shig.ThirdGuest, err = ExcludeUnknownString(props, "thirdGuest"); err != nil {
		return nil, err
	}
	return video, nil
}

type WithUnknown interface {
	GetUnknownProperties() map[string]interface{}
}

func ExcludeUnknownBool(unknownProps map[string]interface{}, key string) (bool, error) {
	value, ok := unknownProps[key].(bool)
	if !ok {
		return false, fmt.Errorf("cannot parsing '%s'", key)
	}
	return value, nil
}

func ExcludeUnknownString(unknownProps map[string]interface{}, key string) (string, error) {
	value, ok := unknownProps[key].(string)
	if !ok {
		return "", fmt.Errorf("cannot parsing '%s'", key)
	}
	return value, nil
}

func ExcludeUnknownUint(unknownProps map[string]interface{}, key string) (uint, error) {
	value, ok := unknownProps[key].(float64)
	if !ok {
		return 0, fmt.Errorf("cannot parsing '%s'", key)
	}
	return uint(value), nil
}

func ExcludeUnknownTime(unknownProps map[string]interface{}, key string) (*time.Time, error) {
	value, ok := unknownProps[key].(*time.Time)
	if !ok {
		return nil, fmt.Errorf("cannot parsing '%s'", key)
	}
	return value, nil
}

func ExcludeUnknownNullTime(unknownProps map[string]interface{}, key string) sql.NullTime {
	value, err := ExcludeUnknownTime(unknownProps, key)
	if err != nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *value, Valid: true}
}

func ExcludeUnknownUuid(unknownProps map[string]interface{}, key string) (*uuid.UUID, error) {
	value, ok := unknownProps[key].(string)
	if !ok {
		return nil, fmt.Errorf("cannot parsing '%s'", key)
	}

	uuidObj, err := uuid.Parse(value)
	if err != nil {
		return nil, fmt.Errorf("is not an uuid '%s'", key)
	}
	return &uuidObj, nil
}
