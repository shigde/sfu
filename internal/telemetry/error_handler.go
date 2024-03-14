package telemetry

import (
	"fmt"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func RecordError(span trace.Span, err error) {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// RecordErrorf formats according to a format specifier and returns the string as a
// value that satisfies error.
func RecordErrorf(span trace.Span, format string, err error) error {
	err = fmt.Errorf("%s: %w", format, err)
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	return err
}
