package message

import (
	"encoding/json"

	"github.com/pion/webrtc/v3"
)

type Sdp struct {
	Number uint32                     `json:"number"`
	SDP    *webrtc.SessionDescription `json:"sdp"`
}

func SdpUnmarshal(sdpData []byte) (*Sdp, error) {
	var newSdp Sdp
	if err := json.Unmarshal(sdpData, &newSdp); err != nil {
		return nil, err
	}
	return &newSdp, nil
}

func SdpMarshal(sdpObj *Sdp) ([]byte, error) {
	data, err := json.Marshal(sdpObj)
	if err != nil {
		return nil, err
	}
	return data, nil
}
