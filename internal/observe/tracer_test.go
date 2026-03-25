package observe

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
)

func TestInitTracer_NoopWhenNoEndpoint(t *testing.T) {
	shutdown, err := InitTracer("test-service", "")
	if err != nil {
		t.Fatalf("InitTracer() error = %v", err)
	}

	// Verify shutdown is callable without error.
	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown() error = %v", err)
	}

	// Verify the global provider is set (noop — creating a span should not panic).
	tp := otel.GetTracerProvider()
	tracer := tp.Tracer("test")
	_, span := tracer.Start(context.Background(), "test-span")
	span.End()
}

func TestInitTracer_NoopShutdownIdempotent(t *testing.T) {
	shutdown, err := InitTracer("test-service", "")
	if err != nil {
		t.Fatalf("InitTracer() error = %v", err)
	}

	// Calling shutdown multiple times should be safe.
	for i := 0; i < 3; i++ {
		if err := shutdown(context.Background()); err != nil {
			t.Fatalf("shutdown() call %d error = %v", i, err)
		}
	}
}

func TestTracer_ReturnsNamedTracer(t *testing.T) {
	// Ensure noop provider is installed.
	_, _ = InitTracer("test-service", "")

	tracer := Tracer("test.component")
	if tracer == nil {
		t.Fatal("Tracer() returned nil")
	}
}
