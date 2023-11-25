package stream

import (
	"context"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

type liveLobbyManager interface {
	CreateLobbyIngressEndpoint(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription) (struct {
		Answer       *webrtc.SessionDescription
		Resource     uuid.UUID
		RtpSessionId uuid.UUID
	}, error)

	InitLobbyEgressEndpoint(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID) (struct {
		Offer        *webrtc.SessionDescription
		Active       bool
		RtpSessionId uuid.UUID
	}, error)

	FinalCreateLobbyEgressEndpoint(ctx context.Context, lobbyId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription) (struct {
		Answer       *webrtc.SessionDescription
		Active       bool
		RtpSessionId uuid.UUID
	}, error)

	LeaveLobby(ctx context.Context, lobbyId uuid.UUID, userId uuid.UUID) (bool, error)

	StartLiveStream(
		ctx context.Context,
		lobbyId uuid.UUID,
		key string,
		rtmpUrl string,
		userId uuid.UUID,
	) error

	StopLiveStream(
		ctx context.Context,
		lobbyId uuid.UUID,
		userId uuid.UUID,
	) error
}
