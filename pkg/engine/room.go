package engine

type Room struct {
	Id               string `json:"Id"`
	peers            []*Peer
	publicStreamRepo *RtpStreamRepository
}
