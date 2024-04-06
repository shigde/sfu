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

	// CreateLobbyIngressEndpoint
	// Deprecated: Because the Endpoint API is getting simpler
	CreateLobbyIngressEndpoint(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription) (struct {
		Answer       *webrtc.SessionDescription
		Resource     uuid.UUID
		RtpSessionId uuid.UUID
	}, error)

	// InitLobbyEgressEndpoint
	// Deprecated: Because the Endpoint API is getting simpler
	InitLobbyEgressEndpoint(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID) (struct {
		Offer        *webrtc.SessionDescription
		Active       bool
		RtpSessionId uuid.UUID
	}, error)

	// FinalCreateLobbyEgressEndpoint
	// Deprecated: Because the Endpoint API is getting simpler
	FinalCreateLobbyEgressEndpoint(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription) (struct {
		Answer       *webrtc.SessionDescription
		Active       bool
		RtpSessionId uuid.UUID
	}, error)

	// CreateMainStreamLobbyEgressEndpoint
	// Deprecated: Because the Endpoint API is getting simpler
	CreateMainStreamLobbyEgressEndpoint(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription) (struct {
		Answer       *webrtc.SessionDescription
		RtpSessionId uuid.UUID
	}, error)

	// CreateLobbyHostPipe
	// Deprecated: Because the Endpoint API is getting simpler
	CreateLobbyHostPipe(ctx context.Context, u uuid.UUID, offer *webrtc.SessionDescription, instanceId uuid.UUID) (struct {
		Answer       *webrtc.SessionDescription
		Resource     uuid.UUID
		RtpSessionId uuid.UUID
	}, error)

	// CreateLobbyHostIngress
	// Deprecated: Because the Endpoint API is getting simpler
	CreateLobbyHostIngress(ctx context.Context, u uuid.UUID, offer *webrtc.SessionDescription, instanceId uuid.UUID) (struct {
		Answer       *webrtc.SessionDescription
		Resource     uuid.UUID
		RtpSessionId uuid.UUID
	}, error)

	// CloseLobbyHostPipe
	// Deprecated: Because the Endpoint API is getting simpler
	CloseLobbyHostPipe(ctx context.Context, u uuid.UUID, id uuid.UUID) (bool, error)
}
