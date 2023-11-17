package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/shigde/sfu/pkg/authentication"
)

type LobbyApi struct {
	UserId    string
	Token     string
	Session   *http.Cookie
	CsrfToken string
	ShigUrl   string
}

func NewLobbyApi(userId string, token string, shigUrl string) *LobbyApi {
	return &LobbyApi{
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

	return &result, nil
}

func (la *LobbyApi) Start(spaceId string, streamId string, bearer string) error {
	resp, err := la.doRequest("POST", fmt.Sprintf("%s/space/%s/stream/%s/live", la.ShigUrl, spaceId, streamId), bearer, nil)
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

func (la *LobbyApi) Status(spaceId string, streamId string, bearer string) (string, error) {
	resp, err := la.doRequest("GET", fmt.Sprintf("%s/space/%s/stream/%s/status", la.ShigUrl, spaceId, streamId), bearer, nil)
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

func (la *LobbyApi) Stop(spaceId string, streamId string, bearer string) error {
	resp, err := la.doRequest("DELETE", fmt.Sprintf("%s/space/%s/stream/%s/live", la.ShigUrl, spaceId, streamId), bearer, nil)
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

func (la *LobbyApi) doRequest(methode string, url string, bearer string, body *bytes.Buffer) (*http.Response, error) {
	c := http.Client{Timeout: time.Duration(1) * time.Second}
	req, err := http.NewRequest(methode, url, body)
	if err != nil {
		return nil, fmt.Errorf("create new request: %w", err)
	}
	req.Header.Add("Accept", `application/json`)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(body.Len()))
	req.Header.Set("Authorization", bearer)

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("doing request: %w", err)
	}

	la.Session = resp.Cookies()[0]
	la.CsrfToken = resp.Header.Get(reqTokenHeaderName)
	return resp, nil
}
