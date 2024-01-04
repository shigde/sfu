package metric

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

var nodeGraph *NodeGraphMetric

type NodeGraphMetric struct {
	node *prometheus.GaugeVec
	edge *prometheus.GaugeVec
}

type GraphNode struct {
	Id                string
	Title             string
	Subtitle          string
	Stream            string
	CurrentTracks     int
	CurrentMainTracks int
}

func GraphNodeUpdate(node GraphNode, tacks int, mainTracks int) GraphNode {
	if nodeGraph != nil {
		ServiceGraphNodeDelete(node)
		node.CurrentTracks = tacks
		node.CurrentMainTracks = mainTracks
		if vec, err := nodeGraph.node.GetMetricWith(prometheus.Labels{
			"id":            node.Id,
			"title":         node.Title,
			"subtitle":      node.Subtitle,
			"stream":        node.Stream,
			"mainstat":      strconv.Itoa(node.CurrentTracks),
			"secondarystat": strconv.Itoa(node.CurrentMainTracks),
		}); err == nil {
			vec.Inc()
		}
	}
	return node
}

func ServiceGraphNodeDelete(node GraphNode) {
	if nodeGraph != nil {
		_ = nodeGraph.node.Delete(prometheus.Labels{
			"id":            node.Id,
			"title":         node.Title,
			"subtitle":      node.Subtitle,
			"stream":        node.Stream,
			"mainstat":      strconv.Itoa(node.CurrentTracks),
			"secondarystat": strconv.Itoa(node.CurrentMainTracks),
		})
	}
}

func ServiceGraphAddEdge(id string, stream string, source string, target string) {
	// inc with bytes received!!
	if nodeGraph != nil {
		if vec, err := nodeGraph.edge.GetMetricWith(prometheus.Labels{
			"Id":     id,
			"source": source,
			"target": target,
			"stream": stream,
		}); err == nil {
			vec.Inc()
		}
	}
}

func ServiceGraphDeleteEdge(id string, stream string, source string, target string) {
	if nodeGraph != nil {
		_ = nodeGraph.edge.Delete(prometheus.Labels{
			"Id":     id,
			"source": source,
			"target": target,
			"stream": stream,
		})
	}
}

func NewServiceGraphMetrics() (*NodeGraphMetric, error) {
	if nodeGraph != nil {
		return nil, errors.New("node graph metric already exists")
	}

	nodeGraph = &NodeGraphMetric{
		node: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shig",
			Name:      "track_traces_node",
			Help:      "edges",
		}, []string{"id", "title", "subtitle", "stream", "mainstat", "secondarystat"}),

		edge: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "shig",
			Name:      "track_traces_edge",
			Help:      "edges",
		}, []string{"id", "source", "target", "stream"}),
	}
	if err := prometheus.Register(nodeGraph.node); err != nil {
		return nil, fmt.Errorf("register nodeGraph node metric: %w", err)
	}
	if err := prometheus.Register(nodeGraph.edge); err != nil {
		return nil, fmt.Errorf("register nodeGraph edge metric: %w", err)
	}
	return nodeGraph, nil
}
