package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/shigde/sfu/internal/stream"
	"github.com/shigde/sfu/pkg/authentication"
)

type LobbyApi struct {
	*Client
	UserId  string
	Token   string
	ShigUrl string
}

func NewLobbyApi(userId string, token string, shigUrl string, opt ...ClientOption) *LobbyApi {
	client := NewClient(opt...)
	return &LobbyApi{
		Client:  client,
		UserId:  userId,
		Token:   token,
		ShigUrl: shigUrl,
	}
}

func (la *LobbyApi) Login() (*authentication.Token, error) {
	loginUrl := fmt.Sprintf("%s/authenticate", la.ShigUrl)
	user := &authentication.User{
		UserId: la.UserId,
		Token:  la.Token,
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

	la.Session.SetBearer("Bearer " + result.JWT)
	return &result, nil
}

func (la *LobbyApi) Start(spaceId string, streamId string, rtmpUrl string, key string) error {
	liveStreamInfo := &stream.LiveStreamInfo{
		RtmpUrl:   rtmpUrl,
		StreamKey: key,
	}
	userJSON, err := json.Marshal(liveStreamInfo)
	if err != nil {
		return fmt.Errorf("creating json object for liveStreamInfo: %w", err)
	}
	body := bytes.NewBuffer(userJSON)

	resp, err := la.doRequest("POST", fmt.Sprintf("%s/space/%s/stream/%s/live", la.ShigUrl, spaceId, streamId), body)
	if err != nil {
		return fmt.Errorf("running start request: %w", err)
	}
	defer resp.Body.Close()

	if resp.Status != "201 Created" {
		return fmt.Errorf("server answer with wrong status code %s", resp.Status)
	}

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response from start req: %w", err)
	}
	return nil
}

func (la *LobbyApi) Status(spaceId string, streamId string) (string, error) {
	resp, err := la.doRequest("GET", fmt.Sprintf("%s/space/%s/stream/%s/status", la.ShigUrl, spaceId, streamId), nil)
	if err != nil {
		return "offline", fmt.Errorf("running status request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "offline", fmt.Errorf("server answer with wrong status code %s", resp.Status)
	}

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return "offline", fmt.Errorf("reading response from status req: %w", err)
	}
	return "offline", nil
}

func (la *LobbyApi) Stop(spaceId string, streamId string) error {
	resp, err := la.doRequest("DELETE", fmt.Sprintf("%s/space/%s/stream/%s/live", la.ShigUrl, spaceId, streamId), nil)
	if err != nil {
		return fmt.Errorf("running stop request: %w", err)
	}
	defer resp.Body.Close()

	if resp.Status != "201 Created" {
		return fmt.Errorf("server answer with wrong status code %s", resp.Status)
	}

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response from stop req: %w", err)
	}
	return nil

}

func (la *LobbyApi) doRequest(methode string, url string, body *bytes.Buffer) (*http.Response, error) {
	c := http.Client{Timeout: time.Duration(1) * time.Second}
	req, err := http.NewRequest(methode, url, body)
	if err != nil {
		return nil, fmt.Errorf("create new request: %w", err)
	}
	req.Header.Add("Accept", `application/json`)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(body.Len()))
	req.Header.Set("Authorization", la.Session.GetBearer())

	req.AddCookie(la.Session.GetCookie())
	req.Header.Set(reqTokenHeaderName, la.Session.GetCsrfToken())

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("doing request: %w", err)
	}

	la.Session.SetCsrfToken(resp.Header.Get(reqTokenHeaderName))
	return resp, nil
}
