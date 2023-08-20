package message

import (
	"encoding/json"
)

type ChannelMsg struct {
	Id   uint64      `json:"id"`
	Data interface{} `json:"data"`
	Type MsgType     `json:"type"`
}
type MsgType int

const (
	OfferMsg MsgType = iota + 1
	AnswerMsg
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
