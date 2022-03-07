package auth

type Principal struct {
	UID       string `json:"uid"`
	SID       string `json:"sid"`
	Publish   bool   `json:"publish"`
	Subscribe bool   `json:"subscribe"`
}

func (principal *Principal) GetUid() string {
	return principal.UID
}

func (principal *Principal) GetSid() string {
	return principal.SID
}

func (principal *Principal) IsPublisher() bool {
	return principal.Publish
}

func (principal *Principal) IsSubscriber() bool {
	return principal.Subscribe
}
