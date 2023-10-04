package webfinger

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/shigde/sfu/internal/activitypub/instance"
)

type Client struct {
	config *instance.FederationConfig
}

func NewClient(config *instance.FederationConfig) *Client {
	return &Client{config}
}
func (c *Client) GetWebfingerLinks(account string) ([]map[string]interface{}, error) {
	type webfingerResponse struct {
		Links []map[string]interface{} `json:"links"`
	}

	account = strings.TrimLeft(account, "@") // remove any leading @
	accountComponents := strings.Split(account, "@")
	fediverseServer := accountComponents[1]

	// HTTPS is required.
	requestURL, err := url.Parse("https://" + fediverseServer)
	if err != nil {
		return nil, fmt.Errorf("unable to parse fediverse server host %s", fediverseServer)
	}

	requestURL.Path = "/.well-known/webfinger"
	query := requestURL.Query()
	query.Add("resource", fmt.Sprintf("acct:%s", account))
	requestURL.RawQuery = query.Encode()

	// Do not support redirects.
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	response, err := client.Get(requestURL.String())
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	var links webfingerResponse
	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(&links); err != nil {
		return nil, err
	}

	return links.Links, nil
}

// MakeWebfingerResponse will create a new Webfinger response.
func (c *Client) MakeWebfingerResponse(account string, inbox string, host string) WebfingerResponse {
	accountIRI := instance.BuildAccountIri(c.config.InstanceUrl, account)
	streamIRI := instance.BuildStreamURLIri(c.config.InstanceUrl)
	return WebfingerResponse{
		Subject: fmt.Sprintf("acct:%s@%s", account, host),
		Aliases: []string{
			accountIRI.String(),
		},
		Links: []Link{
			{
				Rel:  "self",
				Type: "application/activity+json",
				Href: accountIRI.String(),
			},
			{
				Rel:  "https://webfinger.net/rel/profile-page",
				Type: "text/html",
				Href: accountIRI.String(),
			},
			{
				Rel:  "alternate",
				Type: "application/x-mpegURL",
				Href: streamIRI.String(),
			},
		},
	}
}

// MakeWebFingerRequestResponseFromData converts WebFinger data to an easier
// to use model.
func (c *Client) MakeWebFingerRequestResponseFromData(data []map[string]interface{}) WebfingerProfileRequestResponse {
	response := WebfingerProfileRequestResponse{}
	for _, link := range data {
		if link["rel"] == "self" {
			return WebfingerProfileRequestResponse{
				Self: link["href"].(string),
			}
		}
	}

	return response
}
