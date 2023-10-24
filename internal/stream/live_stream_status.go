package stream

type LiveStreamStatus struct {
	Status LiveStreamStatusValue `json:"status"`
}

type LiveStreamStatusValue int

const (
	LiveStreamStatusOnline LiveStreamStatusValue = iota + 1
	LiveStreamStatusOffline
)
