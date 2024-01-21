package message

import "encoding/json"

type Mute struct {
	Mid  string `json:"mid"`
	Mute bool   `json:"mute"`
}

func MuteUnmarshal(data []byte) (*Mute, error) {
	var newMute Mute
	if err := json.Unmarshal(data, &newMute); err != nil {
		return nil, err
	}
	return &newMute, nil
}

func MuteMarshal(muteObj *Mute) ([]byte, error) {
	data, err := json.Marshal(muteObj)
	if err != nil {
		return nil, err
	}
	return data, nil
}
