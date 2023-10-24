package stream

import (
	"context"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v3"
)

type liveLobbyManager interface {
	AccessLobby(ctx context.Context, liveStreamId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription) (struct {
		Answer       *webrtc.SessionDescription
		Resource     uuid.UUID
		RtpSessionId uuid.UUID
	}, error)

	StartListenLobby(ctx context.Context, liveStreamId uuid.UUID, user uuid.UUID) (struct {
		Offer        *webrtc.SessionDescription
		Active       bool
		RtpSessionId uuid.UUID
	}, error)

	ListenLobby(ctx context.Context, liveStreamId uuid.UUID, user uuid.UUID, offer *webrtc.SessionDescription) (struct {
		Answer       *webrtc.SessionDescription
		Active       bool
		RtpSessionId uuid.UUID
	}, error)

	LeaveLobby(ctx context.Context, liveStreamId uuid.UUID, userId uuid.UUID) (bool, error)

	StartLiveStream(
		ctx context.Context,
		liveStreamId uuid.UUID,
		key string,
		rtmpUrl string,
		userId uuid.UUID,
	) error
}
