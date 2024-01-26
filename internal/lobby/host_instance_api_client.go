package lobby

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

type hostInstanceApiClient struct {
	instanceId uuid.UUID
	token      string
	url        *url.URL
}

func NewHostInstanceApiClient(id uuid.UUID, token string, shigUrl *url.URL) *hostInstanceApiClient {
	return &hostInstanceApiClient{
		instanceId: id,
		token:      token,
		url:        shigUrl,
	}
}

func (a *hostInstanceApiClient) PostHostOffer(spaceId string, streamId string, offer *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	body := bytes.NewBuffer([]byte(offer.SDP))
	c := http.Client{Timeout: time.Duration(1) * time.Second}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/space/%s/stream/%s/pipe", a.url.String(), spaceId, streamId), body)
	if err != nil {
		return nil, fmt.Errorf("requesting answer: %w", err)
	}
	req.Header.Add("Accept", `application/sdp`)
	req.Header.Set("Content-Type", "application/sdp")
	req.Header.Set("Content-Length", strconv.Itoa(body.Len()))
	req.Header.Set("Authorization", a.getBearer())

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

func (a *hostInstanceApiClient) getBearer() string {
	return "Bearer " + a.token
}
