package media

import (
	"encoding/json"
	"errors"
	"net/http"
)

const maxPayloadByte = 1048576

var invalidPayload = errors.New("invalid payload")

func getPayload(w http.ResponseWriter, r *http.Request) (*json.Decoder, error) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		return nil, invalidPayload
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxPayloadByte)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec, nil
}
