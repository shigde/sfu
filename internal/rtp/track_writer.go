package rtp

type trackWriter interface {
	Write(b []byte) (int, error)
}
