package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.16.0"
	"google.golang.org/grpc"
)

const (
	clientStartTimeout = 5 * time.Second
)

func NewTracerProvider(ctx context.Context, config *TelemetryConfig) (*trace.TracerProvider, error) {
	var exporter trace.SpanExporter
	var err error

	if config.Enable {
		exporter, err = newTempoExporter(ctx)
		if err != nil {
			return nil, fmt.Errorf("creating OTLP trace exporter: %w", err)
		}
	}

	if !config.Enable {
		exporter, err = newStdoutExporter()
		if err != nil {
			return nil, fmt.Errorf("creating OTLP trace exporter: %w", err)
		}
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String("shig-sfu"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP resource: %w", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithResource(res),
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp, nil
}

func newHttpExporter(ctx context.Context) (*otlptrace.Exporter, error) {
	client := otlptracehttp.NewClient(
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithEndpoint("localhost:3200"),
	)
	exporter, err := otlptrace.New(ctx, client)

	if err != nil {
		return nil, fmt.Errorf("creating HTTP trace exporter: %w", err)
	}
	return exporter, nil
}

func newJaegerExporter() (*jaeger.Exporter, error) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint())
	if err != nil {
		return nil, fmt.Errorf("creating Jaeger trace exporter: %w", err)
	}
	return exp, nil
}

func newStdoutExporter() (*stdouttrace.Exporter, error) {
	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, fmt.Errorf("creating Stdout trace exporter: %w", err)
	}
	return exp, nil
}

func newTempoExporter(ctx context.Context) (*otlptrace.Exporter, error) {
	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint("localhost:4317"),
		otlptracegrpc.WithDialOption(grpc.WithBlock()),
	)
	ctx, cancel := context.WithTimeout(ctx, clientStartTimeout)
	defer cancel()

	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP trace exporter: %w", err)
	}
	return exporter, nil
}
