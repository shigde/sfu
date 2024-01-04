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
	Id         string
	Title      string
	Subtitle   string
	Stream     string // live stream
	Tracks     int
	MainTracks int
}

type GraphEdge struct {
	Id     string
	Source string
	Target string
	Stream string
}

func BuildNode(sessionId string, liveStreamId string, nodeType string) GraphNode {
	return GraphNode{
		Id:         nodeType + "-" + sessionId,
		Title:      nodeType,
		Subtitle:   sessionId,
		Stream:     liveStreamId,
		Tracks:     0,
		MainTracks: 0,
	}
}
func buildEdge(sessionId string, liveStreamId string, nodeType string) GraphEdge {
	endpoint := nodeType + "-" + sessionId
	hub := "hub-" + liveStreamId
	edge := GraphEdge{Id: "edge-" + endpoint, Stream: liveStreamId}
	switch nodeType {
	case "egress":
		edge.Source = hub
		edge.Target = endpoint
	case "ingress":
		edge.Source = endpoint
		edge.Target = hub
	default:
		edge.Target = "unknown"
		edge.Source = "unknown"
	}
	return edge
}

func GraphNodeUpdateInc(node GraphNode, purpose string) GraphNode {
	if nodeGraph != nil {
		GraphNodeDelete(node)
		switch purpose {
		case "main":
			node.MainTracks = node.MainTracks + 1
		default:
			node.Tracks = node.Tracks + 1
		}
		return GraphNodeUpdate(node)
	}
	return node
}

func GraphNodeUpdateDec(node GraphNode, purpose string) GraphNode {
	if nodeGraph != nil {
		GraphNodeDelete(node)
		switch purpose {
		case "main":
			node.MainTracks = node.MainTracks - 1
		default:
			node.Tracks = node.Tracks - 1
		}
		return GraphNodeUpdate(node)
	}
	return node
}

func GraphNodeUpdate(node GraphNode) GraphNode {
	if nodeGraph != nil {
		if vec, err := nodeGraph.node.GetMetricWith(prometheus.Labels{
			"id":            node.Id,
			"title":         node.Title,
			"subtitle":      node.Subtitle,
			"stream":        node.Stream,
			"mainstat":      strconv.Itoa(node.Tracks),
			"secondarystat": strconv.Itoa(node.MainTracks),
		}); err == nil {
			vec.Inc()
		}
	}
	return node
}

func GraphNodeDelete(node GraphNode) {
	if nodeGraph != nil {
		_ = nodeGraph.node.Delete(prometheus.Labels{
			"id":            node.Id,
			"title":         node.Title,
			"subtitle":      node.Subtitle,
			"stream":        node.Stream,
			"mainstat":      strconv.Itoa(node.Tracks),
			"secondarystat": strconv.Itoa(node.MainTracks),
		})
	}
}

func GraphAddEdge(sessionId string, liveStreamId string, nodeType string) {
	edge := buildEdge(sessionId, liveStreamId, nodeType)
	if nodeGraph != nil {
		if vec, err := nodeGraph.edge.GetMetricWith(prometheus.Labels{
			"id":     edge.Id,
			"source": edge.Source,
			"target": edge.Target,
			"stream": edge.Stream,
		}); err == nil {
			vec.Inc()
		}
	}
}

func GraphDeleteEdge(sessionId string, liveStreamId string, nodeType string) {
	edge := buildEdge(sessionId, liveStreamId, nodeType)
	if nodeGraph != nil {
		_ = nodeGraph.edge.Delete(prometheus.Labels{
			"id":     edge.Id,
			"source": edge.Source,
			"target": edge.Target,
			"stream": edge.Stream,
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
