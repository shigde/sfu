package stats

import (
	"strconv"

	"github.com/pion/webrtc/v3"
)

func SSRCtoString(ssrc webrtc.SSRC) string {
	return strconv.Itoa(int(ssrc))
}
