package rtp

import "sync"

type rtpDispatcher struct {
	sync.Mutex
	trackWriter []*trackWriter
}
