package lobby

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/pkg/authentication"
)

type hostInstanceApiClient struct {
	instanceId uuid.UUID
	name       string
	token      string
	url        *url.URL
	bearer     string
}

func NewHostInstanceApiClient(id uuid.UUID, token string, name string, shigUrl *url.URL) *hostInstanceApiClient {
	return &hostInstanceApiClient{
		instanceId: id,
		name:       name,
		token:      token,
		url:        shigUrl,
	}
}

func (a *hostInstanceApiClient) Login() (*authentication.Token, error) {
	loginUrl := fmt.Sprintf("%s/authenticate", a.url.String())
	user := &authentication.User{
		UserId: a.name,
		Token:  a.token,
	}
	userJSON, err := json.Marshal(user)
	if err != nil {
		return nil, fmt.Errorf("creating json object: %w", err)
	}
	body := bytes.NewBuffer(userJSON)

	c := http.Client{Timeout: time.Duration(1) * time.Second}
	req, err := http.NewRequest("POST", loginUrl, body)
	if err != nil {
		return nil, fmt.Errorf("create login request: %w", err)
	}
	req.Header.Add("Accept", `application/json`)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(body.Len()))

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("doing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("server answer with wrong status code %s", resp.Status)
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response from start req: %w", err)
	}

	var result authentication.Token
	err = json.Unmarshal(resBody, &result)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling data from request.")
	}

	a.setBearer("Bearer " + result.JWT)
	return &result, nil
}

func (a *hostInstanceApiClient) PostHostPipeOffer(spaceId string, streamId string, offer *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	requestUrl := fmt.Sprintf("%s/space/%s/stream/%s/pipe", a.url.String(), spaceId, streamId)
	return a.doOfferRequest(requestUrl, offer)
}

func (a *hostInstanceApiClient) PostHostIngressOffer(spaceId string, streamId string, offer *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	requestUrl := fmt.Sprintf("%s/space/%s/stream/%s/hostingress", a.url.String(), spaceId, streamId)
	return a.doOfferRequest(requestUrl, offer)
}

func (a *hostInstanceApiClient) doOfferRequest(reqUrl string, offer *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	body := bytes.NewBuffer([]byte(offer.SDP))
	c := http.Client{Timeout: time.Duration(1) * time.Second}
	req, err := http.NewRequest("POST", reqUrl, body)

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
	return a.bearer
}
func (a *hostInstanceApiClient) setBearer(bearer string) {
	a.bearer = bearer
}
