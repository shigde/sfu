package webfinger

// WebfingerResponse represents a Webfinger response.
type WebfingerResponse struct {
	Aliases []string `json:"aliases"`
	Subject string   `json:"subject"`
	Links   []Link   `json:"links"`
}

// WebfingerProfileRequestResponse represents a Webfinger profile request response.
type WebfingerProfileRequestResponse struct {
	Self string
}

// Link represents a Webfinger response Link entity.
type Link struct {
	Rel  string `json:"rel"`
	Type string `json:"type"`
	Href string `json:"href"`
}
