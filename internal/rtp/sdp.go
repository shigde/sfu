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

		switch strings.TrimSpace(infoString) {
		case "1":
			trackSdpInfo.Purpose = PurposeGuest
		case "2":
			trackSdpInfo.Purpose = PurposeMain
		default:
			trackSdpInfo.Purpose = PurposeGuest
		}

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
					info := sdp.Information(fmt.Sprintf("%d", sdpInfo.Purpose))
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
