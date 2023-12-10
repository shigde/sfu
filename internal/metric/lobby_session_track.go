package metric

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type LabelType string
type Labels map[LabelType]string

const (
	Session   LabelType = "session"
	Stream    LabelType = "stream"
	TrackId   LabelType = "track"
	TrackKind LabelType = "kind"      // video | audio
	TrackType LabelType = "type"      // guest | main
	Direction LabelType = "direction" // ingress | egress
)

var lobbySessionTrackMetric *LobbySessionTrackMetric
var lobbySessionTrackMetricLabels = []string{string(Session), string(Stream), string(TrackId), string(TrackKind), string(TrackType), string(Direction)}

type TrackMetric struct {
	packet          *prometheus.CounterVec
	packetBytes     *prometheus.CounterVec
	nack            *prometheus.CounterVec
	pli             *prometheus.CounterVec
	fir             *prometheus.CounterVec
	packetLossTotal *prometheus.CounterVec
	packetLoss      *prometheus.HistogramVec
	jitter          *prometheus.HistogramVec
	rtt             *prometheus.HistogramVec
}

type LobbySessionTrackMetric struct {
	ingressTracks *TrackMetric
	egressTracks  *TrackMetric
}

func PacketInc(labels Labels, pkg uint64) {
	if trackMetric := chooseDirection(labels); trackMetric != nil {
		trackMetric.packet.With(toPromLabels(labels)).Add(float64(pkg))
	}
}
func PacketDel(labels Labels) {
	if trackMetric := chooseDirection(labels); trackMetric != nil {
		trackMetric.packet.Delete(toPromLabels(labels))
	}
}

func PacketBytesInc(labels Labels, pkg uint64) {
	if trackMetric := chooseDirection(labels); trackMetric != nil {
		trackMetric.packet.With(toPromLabels(labels)).Add(float64(pkg))
	}
}
func PacketBytesDel(labels Labels) {
	if trackMetric := chooseDirection(labels); trackMetric != nil {
		trackMetric.packet.Delete(toPromLabels(labels))
	}
}

func NackInc(labels Labels, nack uint32) {
	if nack == 0 {
		return
	}
	if trackMetric := chooseDirection(labels); trackMetric != nil {
		trackMetric.nack.With(toPromLabels(labels)).Add(float64(nack))
	}
}
func NackDel(labels Labels) {
	if trackMetric := chooseDirection(labels); trackMetric != nil {
		trackMetric.nack.Delete(toPromLabels(labels))
	}
}

func PliInc(labels Labels, pli uint32) {
	if pli == 0 {
		return
	}
	if trackMetric := chooseDirection(labels); trackMetric != nil {
		trackMetric.packet.With(toPromLabels(labels)).Add(float64(pkg))
	}
}
func PliDel(labels Labels) {
	if trackMetric := chooseDirection(labels); trackMetric != nil {
		trackMetric.packet.Delete(toPromLabels(labels))
	}
}

func FirInc(labels Labels, pkg uint64) {
	if trackMetric := chooseDirection(labels); trackMetric != nil {
		trackMetric.packet.With(toPromLabels(labels)).Add(float64(pkg))
	}
}
func FirDel(labels Labels) {
	if trackMetric := chooseDirection(labels); trackMetric != nil {
		trackMetric.packet.Delete(toPromLabels(labels))
	}
}
func PacketLossTotalInc(labels Labels, pkg uint64) {
	if trackMetric := chooseDirection(labels); trackMetric != nil {
		trackMetric.packet.With(toPromLabels(labels)).Add(float64(pkg))
	}
}
func PacketLossTotalDel(labels Labels) {
	if trackMetric := chooseDirection(labels); trackMetric != nil {
		trackMetric.packet.Delete(toPromLabels(labels))
	}
}
func PacketLossInc(labels Labels, pkg uint64) {
	if trackMetric := chooseDirection(labels); trackMetric != nil {
		trackMetric.packet.With(toPromLabels(labels)).Add(float64(pkg))
	}
}
func PacketLossDel(labels Labels) {
	if trackMetric := chooseDirection(labels); trackMetric != nil {
		trackMetric.packet.Delete(toPromLabels(labels))
	}
}
func JitterInc(labels Labels, pkg uint64) {
	if trackMetric := chooseDirection(labels); trackMetric != nil {
		trackMetric.packet.With(toPromLabels(labels)).Add(float64(pkg))
	}
}
func JitterDel(labels Labels) {
	if trackMetric := chooseDirection(labels); trackMetric != nil {
		trackMetric.packet.Delete(toPromLabels(labels))
	}
}

func RttInc(labels Labels, pkg uint64) {
	if trackMetric := chooseDirection(labels); trackMetric != nil {
		trackMetric.packet.With(toPromLabels(labels)).Add(float64(pkg))
	}
}
func RttDel(labels Labels) {
	if trackMetric := chooseDirection(labels); trackMetric != nil {
		trackMetric.packet.Delete(toPromLabels(labels))
	}
}

func newTrackMetric(direction string) (*TrackMetric, error) {
	m := &TrackMetric{
		packet: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "shig",
			Subsystem: direction + "_packet",
			Name:      "total",
		}, lobbySessionTrackMetricLabels),
		packetBytes: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "shig",
			Subsystem: direction + "_packet",
			Name:      "bytes",
		}, lobbySessionTrackMetricLabels),
		nack: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "shig",
			Subsystem: direction + "_nack",
			Name:      "total",
		}, lobbySessionTrackMetricLabels),
		pli: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "shig",
			Subsystem: direction + "_pil",
			Name:      "total",
		}, lobbySessionTrackMetricLabels),
		fir: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "shig",
			Subsystem: direction + "_fir",
			Name:      "total",
		}, lobbySessionTrackMetricLabels),
		packetLossTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "shig",
			Subsystem: direction + "_packet_loss",
			Name:      "total",
		}, lobbySessionTrackMetricLabels),
		packetLoss: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "shig",
			Subsystem: direction + "_packet_loss",
			Name:      "percent",
			Buckets:   []float64{0.0, 0.1, 0.3, 0.5, 0.7, 1, 5, 10, 40, 100},
		}, lobbySessionTrackMetricLabels),
		jitter: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "shig",
			Subsystem: direction + "_jitter",
			Name:      "us",
			Buckets:   []float64{100, 500, 1500, 3000, 6000, 12000, 24000, 48000, 96000, 192000},
		}, lobbySessionTrackMetricLabels),
		rtt: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "shig",
			Subsystem: direction + "_rtt",
			Name:      "ms",
			Buckets:   []float64{50, 100, 150, 200, 250, 500, 750, 1000, 5000, 10000},
		}, lobbySessionTrackMetricLabels),
	}
	if err := prometheus.Register(m.packet); err != nil {
		return nil, fmt.Errorf("register packet metric: %w", err)
	}
	if err := prometheus.Register(m.packetBytes); err != nil {
		return nil, fmt.Errorf("register packet metric: %w", err)
	}
	if err := prometheus.Register(m.nack); err != nil {
		return nil, fmt.Errorf("register nack metric: %w", err)
	}
	if err := prometheus.Register(m.pli); err != nil {
		return nil, fmt.Errorf("register pli metric: %w", err)
	}
	if err := prometheus.Register(m.fir); err != nil {
		return nil, fmt.Errorf("register fir metric: %w", err)
	}
	if err := prometheus.Register(m.packetLossTotal); err != nil {
		return nil, fmt.Errorf("register packetLossTotal metric: %w", err)
	}
	if err := prometheus.Register(m.packetLoss); err != nil {
		return nil, fmt.Errorf("register packetLoss metric: %w", err)
	}

	if err := prometheus.Register(m.jitter); err != nil {
		return nil, fmt.Errorf("register jitter metric: %w", err)
	}
	if err := prometheus.Register(m.rtt); err != nil {
		return nil, fmt.Errorf("register rtt metric: %w", err)
	}

	return m, nil
}

func NewLobbySessionTrackMetrics() (*LobbySessionTrackMetric, error) {
	if lobbySessionTrackMetric != nil {
		return nil, errors.New("lobby session track metric already exists")
	}

	lobbySessionTrackMetric = &LobbySessionTrackMetric{
		runningTracks: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shig",
			Name:      "lobby_session_tracks_",
			Help:      "running lobby session tracks",
		}, []string{"session", "track"}),
	}
	if err := prometheus.Register(lobbySessionTrackMetric.runningTracks); err != nil {
		return nil, fmt.Errorf("register runningLobby metric: %w", err)
	}

	return lobbySessionTrackMetric, nil
}

func chooseDirection(labels Labels) *TrackMetric {
	if lobbySessionTrackMetric == nil {
		return nil
	}
	if labels[Direction] == "ingress" {
		return lobbySessionTrackMetric.ingressTracks
	}

	return lobbySessionTrackMetric.egressTracks
}

func toPromLabels(labels Labels) prometheus.Labels {
	var promLabel prometheus.Labels
	for key, val := range labels {
		promLabel[string(key)] = val
	}
	return promLabel
}
