package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/pion/webrtc/v3"
)

type Whip struct {
}

func NewWhip() *Whip {
	return &Whip{}
}

func (w *Whip) GetAnswer(spaceId string, streamId string, bearer string, offer *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	body := bytes.NewBuffer([]byte(offer.SDP))
	c := http.Client{Timeout: time.Duration(1) * time.Second}
	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:8080/space/%s/stream/%s/whip", spaceId, streamId), body)
	if err != nil {
		return nil, fmt.Errorf("requesting answer: %w", err)
	}
	req.Header.Add("Accept", `application/sdp`)
	req.Header.Set("Content-Type", "application/sdp")
	req.Header.Set("Content-Length", strconv.Itoa(body.Len()))
	req.Header.Set("Authorization", bearer)

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requesting answer: %w", err)
	}
	defer resp.Body.Close()

	if resp.Status != "201 Created" {
		return nil, fmt.Errorf("server answer with wrong status code %s", resp.Status)
	}

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading sdp answer from response body: %w", err)
	}

	answer := string(response)
	return &webrtc.SessionDescription{SDP: answer, Type: webrtc.SDPTypeAnswer}, nil
}
