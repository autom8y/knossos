package observe

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/autom8y/knossos/internal/serve"
)

// OTELMiddleware returns HTTP tracing middleware that creates a root span
// per HTTP request and records standard HTTP attributes.
func OTELMiddleware() serve.Middleware {
	tracer := Tracer("clew.http")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
			ctx, span := tracer.Start(r.Context(), spanName,
				trace.WithSpanKind(trace.SpanKindServer),
			)
			defer span.End()

			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.route", r.URL.Path),
			)

			// Wrap the writer to capture the status code.
			rec := &statusCapture{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rec, r.WithContext(ctx))

			span.SetAttributes(
				attribute.Int("http.status_code", rec.statusCode),
			)
		})
	}
}

// statusCapture wraps http.ResponseWriter to capture the written status code.
type statusCapture struct {
	http.ResponseWriter
	statusCode int
}

func (sc *statusCapture) WriteHeader(code int) {
	sc.statusCode = code
	sc.ResponseWriter.WriteHeader(code)
}
