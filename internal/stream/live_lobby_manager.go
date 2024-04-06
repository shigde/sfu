package stream

import (
	"context"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
	"github.com/shigde/sfu/internal/lobby/resources"
)

type liveLobbyManager interface {
	NewIngressResource(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription, option ...resources.Option) (*resources.WebRTC, error)
	NewEgressResource(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription, option ...resources.Option) (*resources.WebRTC, error)
	LeaveLobby(ctx context.Context, lobbyId uuid.UUID, userId uuid.UUID) (bool, error)

	// Live Stream Publishing API

	StartLiveStream(ctx context.Context, lobbyId uuid.UUID, key string, rtmpUrl string, userId uuid.UUID) error
	StopLiveStream(ctx context.Context, lobbyId uuid.UUID, userId uuid.UUID) error

	// Deprecated API

	// CreateLobbyIngressEndpoint deprecated
	CreateLobbyIngressEndpoint(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription) (struct {
		Answer       *webrtc.SessionDescription
		Resource     uuid.UUID
		RtpSessionId uuid.UUID
	}, error)

	// InitLobbyEgressEndpoint deprecated
	InitLobbyEgressEndpoint(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID) (struct {
		Offer        *webrtc.SessionDescription
		Active       bool
		RtpSessionId uuid.UUID
	}, error)

	// FinalCreateLobbyEgressEndpoint deprecated
	FinalCreateLobbyEgressEndpoint(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription) (struct {
		Answer       *webrtc.SessionDescription
		Active       bool
		RtpSessionId uuid.UUID
	}, error)

	// CreateMainStreamLobbyEgressEndpoint deprecated
	CreateMainStreamLobbyEgressEndpoint(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription) (struct {
		Answer       *webrtc.SessionDescription
		RtpSessionId uuid.UUID
	}, error)

	// CreateLobbyHostPipe deprecated
	CreateLobbyHostPipe(ctx context.Context, u uuid.UUID, offer *webrtc.SessionDescription, instanceId uuid.UUID) (struct {
		Answer       *webrtc.SessionDescription
		Resource     uuid.UUID
		RtpSessionId uuid.UUID
	}, error)

	// CreateLobbyHostIngress deprecated
	CreateLobbyHostIngress(ctx context.Context, u uuid.UUID, offer *webrtc.SessionDescription, instanceId uuid.UUID) (struct {
		Answer       *webrtc.SessionDescription
		Resource     uuid.UUID
		RtpSessionId uuid.UUID
	}, error)

	// CloseLobbyHostPipe deprecated
	CloseLobbyHostPipe(ctx context.Context, u uuid.UUID, id uuid.UUID) (bool, error)
}
