package rtp

type VideoQuality int32

const (
	VideoQuality_LOW    VideoQuality = 0
	VideoQuality_MEDIUM VideoQuality = 1
	VideoQuality_HIGH   VideoQuality = 2
	VideoQuality_OFF    VideoQuality = 3
)

// Enum value maps for VideoQuality.
var (
	VideoQuality_name = map[int32]string{
		0: "LOW",
		1: "MEDIUM",
		2: "HIGH",
		3: "OFF",
	}
	VideoQuality_value = map[string]int32{
		"LOW":    0,
		"MEDIUM": 1,
		"HIGH":   2,
		"OFF":    3,
	}
)
