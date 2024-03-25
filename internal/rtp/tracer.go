package rtp

import (
	"context"

	"github.com/shigde/sfu/internal/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = telemetry.TracerName

func newTraceSpan(ctx context.Context, sessionCxt context.Context, spanName string) (context.Context, trace.Span) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, spanName, trace.WithAttributes(
		attribute.String("sessionId", sessionCxt.Value("sessionId").(string)),
		attribute.String("liveStreamId", sessionCxt.Value("liveStreamId").(string)),
		attribute.String("userId", sessionCxt.Value("userId").(string)),
	))
	return ctx, span
}
