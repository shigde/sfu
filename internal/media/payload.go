package media

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/pion/webrtc/v3"
)

const maxPayloadByte = 1048576

var (
	invalidContentType = errors.New("invalid content type")
	invalidPayload     = errors.New("invalid payload")
	emptyPayload       = errors.New("empty payload")
)

func getJsonPayload(w http.ResponseWriter, r *http.Request) (*json.Decoder, error) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		return nil, invalidContentType
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxPayloadByte)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec, nil
}

func getSdpPayload(w http.ResponseWriter, r *http.Request, sdpType webrtc.SDPType) (*webrtc.SessionDescription, error) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/sdp" {
		return nil, invalidContentType
	}
	if r.Body == nil {
		return nil, emptyPayload
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxPayloadByte)
	bodyBytes, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		return nil, invalidPayload
	}
	bodyString := string(bodyBytes)

	return &webrtc.SessionDescription{
		SDP:  bodyString,
		Type: sdpType,
	}, nil
}
