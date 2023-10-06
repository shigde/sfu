package models

import (
	"database/sql"
	"net/url"

	"gorm.io/gorm"
)

type Video struct {
	Iri             string       `gorm:"not null;index;unique;"`
	Uuid            string       `gorm:"not null;index;unique;"`
	Name            string       `gorm:""`
	ShigActive      bool         `gorm:"not null;default:false;"`
	OwnerId         uint         `gorm:"not null;"`
	Owner           *Actor       `gorm:"foreignKey:OwnerId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	ChannelId       uint         `gorm:"not null;"`
	Channel         *Actor       `gorm:"foreignKey:ChannelId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Guests          []*Actor     `gorm:"many2many:video_guests;"`
	IsLiveBroadcast bool         `gorm:"not null;default:false;"`
	LiveSaveReplay  bool         `gorm:"not null;default:false;"`
	PermanentLive   bool         `gorm:""`
	LatencyMode     uint         `gorm:"not null;default:1;"`
	Published       sql.NullTime `gorm:""`
	State           uint         `gorm:"not null;default:0"`
	iriUrl          *url.URL     `gorm:"-"`
	gorm.Model
}

func (s *Video) GetVideoIri() *url.URL {
	if s.iriUrl == nil {
		s.iriUrl, _ = url.Parse(s.Iri)
	}
	return s.iriUrl
}

func (s *Video) GetState() VideoState {
	return VideoState(s.State)
}

func (s *Video) GetLatencyMode() LiveVideoLatencyMode {
	return LiveVideoLatencyMode(s.LatencyMode)
}

type VideoState uint

const (
	PUBLISHED VideoState = iota
	TO_TRANSCODE
	TO_IMPORT
	WAITING_FOR_LIVE
	LIVE_ENDED
	TO_MOVE_TO_EXTERNAL_STORAGE
	TRANSCODING_FAILED
	TO_MOVE_TO_EXTERNAL_STORAGE_FAILED
	TO_EDIT
)

type LiveVideoLatencyMode uint

const (
	DEFAULT LiveVideoLatencyMode = iota
	HIGH_LATENCY
	SMALL_LATENCY
)
