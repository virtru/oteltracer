# Virtru Open Telemetry Tracer helper library

A basic wrapper to make it easy to decorate Golang HTTP servers with OpenTelemetry traces and context propagators.

This wrapper is designed to write to a Collector/Forwarder service that supports OpenTelemetry traces.

# Usage

> NOTE: This library writes to a collector endpoint defined by the env var `OTLP_COLLECTOR_ENDPOINT` - ex. localhost:4345 - this env var must be set in your app's environment, and it must point towards a correctly-configured Collector/Forwarder instance.


1. Import this library in your Go server `main` routine:

    ``` go
    import (
        "github.com/virtru/oteltracer"
    )
    ```

2. Initialize and register a global tracer instance for your app (typically in your server main loop) - this is what sets up the gRPC connection to the collector:

    ``` go

    tracerCancel, err := oteltracer.InitTracer("my-app")
    if err != nil {
        logger.Fatal("Error initializing tracer")
    }
    defer tracerCancel()
    ```

3. Now, anywhere and everywhere you want to add a Span or a Trace (in the same module, or elsewhere in your app), just:

    ``` go
    // This is the only module you should need to import for basic Spans, there are other
    // helper libraries for e.g. mux Middleware or HTTP handlers if you want them.
    import go.opentelemetry.io/otel

    ...
    ...

    childCtx, myOpSpan := otel.Tracer("this go module name").Start(ctx, "my operation")
    defer myOpSpan.End() // Spans should be Ended when done.
    ```

    you can create as many Spans as you want, and calling `otel.Tracer("my-module-name")` will get an instance of the previously-created global tracer you initialized (if there is one, if there is not then a no-op tracer will be created and used)

4. Note that creating Spans requires you to provide a vanilla [Go Context object](https://pkg.go.dev/context), and returns a `child` Context object that is linked to the parent Context. So you can nest Spans (as deeply as you wish) by calling Tracer.Start and passing in the "child" Context created by the parent Tracer.Start:

    ``` go
    // This is the only module you should need to import for basic Spans, there are other
    // helper libraries for e.g. mux Middleware or HTTP handlers if you want them.
    import go.opentelemetry.io/otel

    ...
    ...

    childCtx, myOpSpan := otel.Tracer("this go module name").Start(ctx, "my operation")
    defer myOpSpan.End()  // Spans should be Ended when done.

    doMyOp()

    grandChildContext, anotherOpSpan := otel.Tracer("this go module name").Start(childCtx, "another operation")
    defer anotherOpSpan.End() // Spans should be Ended when done.

    doAnotherOp()

    ```

    Since you create Spans "from" Context objects, and the Context object itself is what "carries" Span parent-child links, you can [pass Contexts around in the normative, recommended Go way](https://pkg.go.dev/context) and just create Spans from nested Contexts wherever you need to.

5. Alternate example: Use it an a garden-variety HTTP handler, getting the root Context from the Go HttpRequest object.

    ``` go

    // This is the only module you should need to import for basic Spans, there are other
    // helper libraries for e.g. mux Middleware or HTTP handlers if you want them.
    import go.opentelemetry.io/otel

    ...
    ...

    var tracer = otel.Tracer("my-handlers-module")

    ...
    ...

    // healthz is a liveness probe.
    func getHealthzHandler() http.Handler {
        healthzHandler := func(w http.ResponseWriter, req *http.Request) {
            // Obtain a Golang Context object from the request - if the incoming request
            // has otel request tracing headers on it, those will be propagated/linked
            // with whatever uses this Context object as a starting point.
            reqCtx := req.Context()
            // This will get the global registered Tracer instance that `otelTracer.InitTracer("name")`
            // initialized and registered, and start a Span.
            // You can create as many Traces/Spans as you want in your code by calling
            // `otel.Tracer("tracerName").Start(ctx, "spanName")` - as long as they're all derived
            // from the same root Context(), they will be linked.
            childCtx, span := tracer.Start(reqCtx, "healthzHandler")
            // Note that at this point you could pass `childCtx` to inner functions, etc
            // as a 'starting point' for creating additional nested Spans.
            defer span.End() // Spans should be Ended when done.
            span.AddEvent("Healthz request received")
            // log.Printf("Healthz request received, ready status is: %v", isReady)
            if !isReady {
                http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
                return
            }
            w.WriteHeader(http.StatusOK)
        }

        return http.HandlerFunc(healthzHandler)
    }
    ```

    NOTE that there is an upstream OpenTelemetry Mux middleware that probably makes this less verbose and easier: https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux

## Example Collector

To deploy your own OpenTelemetry Collector for local testing, you can use the upstream Helm chart:

https://github.com/open-telemetry/opentelemetry-helm-charts/tree/main/charts/opentelemetry-collector

A sample `values.yaml` that configures the Collector chart to forward to a DataDog backend can be found in [example/collector](example/collector)


To deploy the Collector with those sample values to a K8S cluster:

1. Create a K8S Secret named `otel-datadog-secrets` with your DataDog API key: `kubectl create secret generic otel-datadog-secrets --from-literal=DD_API_KEY='<SECRET>'`
1. Add the Otel chart repo: `helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts`
1. Install the chart with the sample value overrides: `helm install otel open-telemetry/opentelemetry-collector -f example/collector/values.yaml`

Generally, you do **not** need to deploy a Collector with your app - one Collector would be deployed per-cluster as a platform service - the above is provided as an example for local testing and experimentation.

## Versioning/releases/changelog

See CONTRIBUTING.md
