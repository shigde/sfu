package rtp

import (
	"context"

	"github.com/shigde/sfu/internal/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = telemetry.TracerName

func newTraceSpan(ctx context.Context, spanName string, sessionId string, lobbyId string) (context.Context, trace.Span) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, spanName, trace.WithAttributes(
		attribute.String("sessionId", sessionId),
		attribute.String("lobbyId", lobbyId),
	))
	return ctx, span
}