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

type Whep struct {
	*Client
}

func NewWhep(opt ...ClientOption) *Whep {
	client := NewClient(opt...)
	return &Whep{
		client,
	}
}

func (w *Whep) GetOffer(spaceId string, streamId string) (*webrtc.SessionDescription, error) {
	c := http.Client{Timeout: time.Duration(1) * time.Second}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/space/%s/stream/%s/whep", w.url.String(), spaceId, streamId), nil)
	if err != nil {
		return nil, fmt.Errorf("requesting answer: %w", err)
	}
	req.Header.Add("Accept", `application/sdp`)
	req.Header.Set("Content-Type", "application/sdp")
	req.Header.Set("Content-Length", strconv.Itoa(0))
	req.Header.Set("Authorization", w.Session.GetBearer())

	req.AddCookie(w.Session.GetCookie())
	req.Header.Set(reqTokenHeaderName, w.Session.GetCsrfToken())

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requesting answer: %w", err)
	}
	defer resp.Body.Close()

	w.Session.SetCsrfToken(resp.Header.Get(reqTokenHeaderName))

	if resp.Status != "201 Created" {
		return nil, fmt.Errorf("server answer with wrong status code %s", resp.Status)
	}

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading sdp answer from response body: %w", err)
	}

	offer := string(response)
	return &webrtc.SessionDescription{SDP: offer, Type: webrtc.SDPTypeOffer}, nil
}

func (w *Whep) SendAnswer(spaceId string, streamId string, bearer string, answer *webrtc.SessionDescription) error {
	body := bytes.NewBuffer([]byte(answer.SDP))
	c := http.Client{Timeout: time.Duration(1) * time.Second}
	req, err := http.NewRequest("PATCH", fmt.Sprintf("http://localhost:8080/space/%s/stream/%s/whep", spaceId, streamId), body)
	if err != nil {
		return fmt.Errorf("sending answer: %w", err)
	}
	req.Header.Add("Accept", `application/sdp`)
	req.Header.Set("Content-Type", "application/sdp")
	req.Header.Set("Content-Length", strconv.Itoa(body.Len()))
	req.Header.Set("Authorization", bearer)

	req.AddCookie(w.Session.GetCookie())
	req.Header.Set(reqTokenHeaderName, w.Session.GetCsrfToken())

	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("requesting answer: %w", err)
	}
	defer resp.Body.Close()

	w.Session.SetCsrfToken(resp.Header.Get(reqTokenHeaderName))

	if resp.Status != "201 Created" {
		return fmt.Errorf("server answer with wrong status code %s", resp.Status)
	}
	return nil
}
