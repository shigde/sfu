package rtp

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

func getTrackInfo(sdp webrtc.SessionDescription, session uuid.UUID) (map[string]*TrackInfo, error) {
	trackInfoMap := make(map[string]*TrackInfo)
	sdpObj, err := sdp.Unmarshal()
	if err != nil {
		return nil, fmt.Errorf("unmarshal sdp: %w", err)
	}
	for _, desc := range sdpObj.MediaDescriptions {
		if desc.MediaName.Media != "video" && desc.MediaName.Media != "audio" {
			continue
		}

		trackInfo := &TrackInfo{
			SessionId: session,
		}
		infoString := ""
		if desc.MediaTitle != nil {
			infoString = desc.MediaTitle.String()
		}

		switch strings.TrimSpace(infoString) {
		case "1":
			trackInfo.Kind = TrackInfoKindGuest
		case "2":
			trackInfo.Kind = TrackInfoKindStream
		default:
			trackInfo.Kind = TrackInfoKindGuest
		}
		msid, fund := desc.Attribute("msid")
		if fund {
			trackInfoMap[msid] = trackInfo
		}
	}
	return trackInfoMap, nil
}
