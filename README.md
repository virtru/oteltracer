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

## Example Collector

To deploy your own OpenTelemetry Collector, you can use the upstream Helm chart:

https://github.com/open-telemetry/opentelemetry-helm-charts/tree/main/charts/opentelemetry-collector

A sample `values.yaml` that configures the Collector chart to forward to a DataDog backend can be found in [example/collector](example/collector)


To deploy the Collector with those sample values to a K8S cluster:

1. Create a K8S Secret named `otel-datadog-secrets` with your DataDog API key: `kubectl create secret generic otel-datadog-secrets --from-literal=DD_API_KEY='<SECRET>'`
1. Add the Otel chart repo: `helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts`
1. Install the chart with the sample value overrides: `helm install otel open-telemetry/opentelemetry-collector -f example/collector/values.yaml`

## Versioning/releases/changelog

See CONTRIBUTING.md
