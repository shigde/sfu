package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	TracerName         = "github.com/shigde/sfu"
	SessionIdCtxKey    = "sessionId"
	UserIdCtxKey       = "liveStreamId"
	LiveStreamIdCtxKey = "userId"
)

func NewTraceSpan(ctx context.Context, sessionCtx context.Context, spanName string) (context.Context, trace.Span) {
	return NewTraceSpanWithTracer(TracerName, ctx, sessionCtx, spanName)
}

func NewTraceSpanWithTracer(tracer string, ctx context.Context, sessionCtx context.Context, spanName string) (context.Context, trace.Span) {
	ctx, span := otel.Tracer(tracer).Start(ctx, spanName, trace.WithAttributes(
		attribute.String(SessionIdCtxKey, sessionCtx.Value(SessionIdCtxKey).(string)),
		attribute.String(LiveStreamIdCtxKey, sessionCtx.Value(LiveStreamIdCtxKey).(string)),
		attribute.String(UserIdCtxKey, sessionCtx.Value(UserIdCtxKey).(string)),
	))
	return ctx, span
}

func CopySessionContext(sessionCtx context.Context) context.Context {
	return ContextWithSessionValue(
		context.Background(),
		sessionCtx.Value(SessionIdCtxKey).(string),
		sessionCtx.Value(LiveStreamIdCtxKey).(string),
		sessionCtx.Value(UserIdCtxKey).(string),
	)
}

func ContextWithSessionValue(ctx context.Context, sessionId string, liveStreamId string, userId string) context.Context {
	ctx = context.WithValue(ctx, SessionIdCtxKey, sessionId)
	ctx = context.WithValue(ctx, LiveStreamIdCtxKey, liveStreamId)
	ctx = context.WithValue(ctx, UserIdCtxKey, userId)
	return ctx
}
