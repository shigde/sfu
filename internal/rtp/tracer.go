package rtp

import (
	"context"

	"github.com/shigde/sfu/internal/telemetry"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = telemetry.TracerName

func newTraceSpan(ctx context.Context, sessionCtx context.Context, spanName string) (context.Context, trace.Span) {
	return telemetry.NewTraceSpanWithTracer(tracerName, ctx, sessionCtx, spanName)
}
