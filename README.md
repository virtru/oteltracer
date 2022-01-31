# Virtru Open Telemetry Tracer helper library

A basic wrapper to make it easy to decorate Golang HTTP servers with OpenTelemetry traces and context propagators.

This wrapper is designed to write to a Collector/Forwarder service that supports OpenTelemetry traces.

# Usage

This library writes to a collector endpoint defined by the env var `OTLP_COLLECTOR_ENDPOINT` - ex. localhost:4345

Import the library in your Go server, and a few otel support libraries:

``` go
import (
    "github.com/virtru-corp/oteltracer"
    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
    "go.opentelemetry.io/otel/trace"
)
```

Initialize it (typically in your server main loop):

``` go

tracerCancel := oteltracer.InitTracer("tdf-proxy")
defer tracerCancel()
```

Use it an a garden-variety handler:

``` go
// healthz is a liveness probe.
func getHealthzHandler() http.Handler {
    healthzHandler := func(w http.ResponseWriter, req *http.Request) {
        ctx := req.Context()
        span := trace.SpanFromContext(ctx)
        span.AddEvent("Healthz request received")
        // log.Printf("Healthz request received, ready status is: %v", isReady)
        if !isReady {
            http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
            return
        }
        w.WriteHeader(http.StatusOK)
    }

    return otelhttp.NewHandler(http.HandlerFunc(healthzHandler), "healthzHandler")
}
```

## Versioning/releases/changelog

See CONTRIBUTING.md
