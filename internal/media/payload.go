package media

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/pion/webrtc/v3"
)

const maxPayloadByte = 1048576

var invalidPayload = errors.New("invalid payload")

func getJsonPayload(w http.ResponseWriter, r *http.Request) (*json.Decoder, error) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		return nil, invalidPayload
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxPayloadByte)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec, nil
}

func getSdpPayload(w http.ResponseWriter, r *http.Request, sdpType webrtc.SDPType) (*webrtc.SessionDescription, error) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/sdp" {
		return nil, invalidPayload
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxPayloadByte)
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("read payload boddy")
	}
	bodyString := string(bodyBytes)

	return &webrtc.SessionDescription{
		SDP:  bodyString,
		Type: sdpType,
	}, nil
}
