package message

import (
	"encoding/json"
)

type ChannelMsg struct {
	Id   int         `json:"id"`
	Data interface{} `json:"data"`
	Type msgType     `json:"type"`
}
type msgType int

const (
	offer msgType = iota + 1
	answer
)

func Unmarshal(rawChannelMsg []byte) (*ChannelMsg, error) {
	var newChannelMsg ChannelMsg
	if err := json.Unmarshal(rawChannelMsg, &newChannelMsg); err != nil {
		return nil, err
	}
	return &newChannelMsg, nil
}

func Marshal(channelMsg *ChannelMsg) ([]byte, error) {
	data, err := json.Marshal(channelMsg)
	if err != nil {
		return nil, err
	}
	return data, nil
}
