package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	otrace "go.opentelemetry.io/otel/trace"
)

var Tracer otrace.Tracer

func InitTracer(serviceName string) (*trace.TracerProvider, error) {
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	Tracer = otel.Tracer(serviceName)
	return tp, nil
}

func ShutdownTracer(ctx context.Context, tp *trace.TracerProvider) {
	if tp != nil {
		if err := tp.Shutdown(ctx); err != nil {
			fmt.Printf("Error shutting down tracer provider: %v\n", err)
		}
	}
}
