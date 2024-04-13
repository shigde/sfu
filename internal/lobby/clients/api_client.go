package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/pkg/authentication"
)

type ApiClient struct {
	login    loginGetter
	spaceId  string
	streamId string
	url      string
	bearer   string
}

type loginGetter interface {
	GetUser() *authentication.User
}

func NewApiClient(loginGetter loginGetter, shigUrl string, spaceId string, streamId string) *ApiClient {
	return &ApiClient{
		login:    loginGetter,
		url:      shigUrl,
		spaceId:  spaceId,
		streamId: streamId,
	}
}

func (a *ApiClient) Login() (*authentication.Token, error) {
	loginUrl := fmt.Sprintf("%s/authenticate", a.url)
	user := a.login.GetUser()
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

func (a *ApiClient) PostWhepOffer(offer *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	requestUrl := fmt.Sprintf("%s/fed/space/%s/stream/%s/whep", a.url, a.spaceId, a.streamId)
	return a.doOfferRequest(requestUrl, offer)
}

func (a *ApiClient) PostWhipOffer(offer *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	requestUrl := fmt.Sprintf("%s/fed/space/%s/stream/%s/whip", a.url, a.spaceId, a.streamId)
	return a.doOfferRequest(requestUrl, offer)
}

func (a *ApiClient) doOfferRequest(reqUrl string, offer *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
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

func (a *ApiClient) getBearer() string {
	return a.bearer
}
func (a *ApiClient) setBearer(bearer string) {
	a.bearer = bearer
}
