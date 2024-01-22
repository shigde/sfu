package rtp

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/pion/sdp/v3"
	"github.com/pion/webrtc/v3"
)

func getIngressTrackSdpInfo(sdp webrtc.SessionDescription, sessionId uuid.UUID, rep *trackSdpInfoRepository) error {
	sdpObj, err := sdp.Unmarshal()
	if err != nil {
		return fmt.Errorf("unmarshal sdp: %w", err)
	}
	for _, desc := range sdpObj.MediaDescriptions {
		if desc.MediaName.Media != "video" && desc.MediaName.Media != "audio" {
			continue
		}

		trackSdpInfo := newTrackSdpInfo(sessionId)
		infoString := ""
		if desc.MediaTitle != nil {
			infoString = desc.MediaTitle.String()
		}

		parseSdpInformation(infoString, trackSdpInfo)

		if msid, fund := desc.Attribute("msid"); fund {
			if idList := strings.SplitAfter(msid, " "); len(idList) == 2 {
				ingressTrackId := idList[1]
				trackSdpInfo.IngressTrackId = ingressTrackId
				rep.Set(trackSdpInfo.Id, trackSdpInfo)
			}
		}
	}
	return nil
}

func MarkStreamAsMain(sdpOrigin *webrtc.SessionDescription, streamID string) (*webrtc.SessionDescription, error) {
	sdpObj, err := sdpOrigin.Unmarshal()
	if err != nil {
		return nil, fmt.Errorf("unmarshal sdp: %w", err)
	}

	for _, desc := range sdpObj.MediaDescriptions {
		if desc.MediaName.Media != "video" && desc.MediaName.Media != "audio" {
			continue
		}
		msid, fund := desc.Attribute("msid")
		if fund && strings.Contains(msid, streamID) {
			info := sdp.Information(fmt.Sprintf("%d", PurposeMain))
			desc.MediaTitle = &info
		}
	}
	sdpByt, err := sdpObj.Marshal()
	if err != nil {
		return nil, fmt.Errorf("marshal new sdp: %w", err)
	}
	sdpNew := &webrtc.SessionDescription{
		Type: sdpOrigin.Type,
		SDP:  string(sdpByt),
	}
	return sdpNew, nil
}

func setEgressTrackInfo(sdpOrigin *webrtc.SessionDescription, repo *trackSdpInfoRepository) (*webrtc.SessionDescription, error) {
	sdpObj, err := sdpOrigin.Unmarshal()
	if err != nil {
		return nil, fmt.Errorf("unmarshal sdp: %w", err)
	}

	for _, desc := range sdpObj.MediaDescriptions {
		if desc.MediaName.Media != "video" && desc.MediaName.Media != "audio" {
			continue
		}

		if msid, fund := desc.Attribute("msid"); fund {
			if idList := strings.SplitAfter(msid, " "); len(idList) == 2 {
				egressTrackId := idList[1]
				if sdpInfo, ok := repo.getSdpInfoByEgressTrackId(egressTrackId); ok {
					info := buildSdpInformation(sdpInfo)
					desc.MediaTitle = &info
					if mid, fundMid := desc.Attribute("mid"); fundMid {
						sdpInfo.EgressMid = mid
					}
				}
			}
		}
	}
	sdpByt, err := sdpObj.Marshal()
	if err != nil {
		return nil, fmt.Errorf("marshal new sdp: %w", err)
	}
	sdpNew := &webrtc.SessionDescription{
		Type: sdpOrigin.Type,
		SDP:  string(sdpByt),
	}
	return sdpNew, nil
}

func parseSdpInformation(infoString string, trackSdpInfo *TrackSdpInfo) {
	if info := strings.SplitAfter(infoString, " "); len(info) == 3 {
		switch strings.TrimSpace(info[0]) {
		case "1":
			trackSdpInfo.Purpose = PurposeGuest
		case "2":
			trackSdpInfo.Purpose = PurposeMain
		default:
			trackSdpInfo.Purpose = PurposeGuest
		}

		switch strings.TrimSpace(info[1]) {
		case "1":
			trackSdpInfo.Mute = true
		case "2":
			trackSdpInfo.Mute = false
		default:
			trackSdpInfo.Mute = false
		}
		trackSdpInfo.Info = info[2]
	}
}

func buildSdpInformation(trackSdpInfo *TrackSdpInfo) sdp.Information {
	muted := "2"
	if trackSdpInfo.Mute {
		muted = "1"
	}
	return sdp.Information(fmt.Sprintf("%d %s %s", trackSdpInfo.Purpose, muted, trackSdpInfo.Info))
}
