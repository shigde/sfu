package rtp

type SimulcastLayer struct {
	Quality VideoQuality `json:"quality,omitempty"`
	Width   uint32       `json:"width,omitempty"`
	Height  uint32       `json:"height,omitempty"`
	// target bitrate in bit per second (bps), server will measure actual
	Bitrate uint32 `json:"bitrate,omitempty"`
	Ssrc    uint32 `json:"ssrc,omitempty"`
}

func (sl *SimulcastLayer) Reset() {
	*sl = SimulcastLayer{}
}

func (sl *SimulcastLayer) GetQuality() VideoQuality {
	if sl != nil {
		return sl.Quality
	}
	return VideoQuality_LOW
}

func (sl *SimulcastLayer) GetWidth() uint32 {
	if sl != nil {
		return sl.Width
	}
	return 0
}

func (sl *SimulcastLayer) GetHeight() uint32 {
	if sl != nil {
		return sl.Height
	}
	return 0
}

func (sl *SimulcastLayer) GetBitrate() uint32 {
	if sl != nil {
		return sl.Bitrate
	}
	return 0
}

func (sl *SimulcastLayer) GetSsrc() uint32 {
	if sl != nil {
		return sl.Ssrc
	}
	return 0
}
