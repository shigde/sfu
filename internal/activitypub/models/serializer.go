package models

import (
	"encoding/json"

	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
)

func Serialize(obj vocab.Type) ([]byte, error) {
	var jsonmap map[string]interface{}
	jsonmap, _ = streams.Serialize(obj)
	b, err := json.Marshal(jsonmap)

	return b, err
}
