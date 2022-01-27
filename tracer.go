package oteltracer

import (
	"context"
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/propagation"
	exporter "go.opentelemetry.io/otel/sdk/export/trace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"

	"google.golang.org/grpc"
)

var otelGRPCCollector = os.Getenv("OTLP_COLLECTOR_ENDPOINT") //ex. localhost:4345

var stdoutTrace = false

func InitTracer(serviceName string) func() {
	var exporter exporter.SpanExporter
	var shutdownFunc func()
	var err error
	if stdoutTrace {
		// Create stdout exporter to be able to retrieve
		// the collected spans.
		log.Println("Starting OpenTelemetry tracer: stdout")
		exporter, err = stdout.NewExporter()
		if err != nil {
			log.Fatal(err)
		}

		shutdownFunc = func() {
			log.Printf("Shutting down stdout tracer")
		}
	} else {

		log.Printf("Starting OpenTelemetry tracer: otlp, configured with endpoint: %s", otelGRPCCollector)
		otlpTracerCtx := context.Background()

		// If the OpenTelemetry Collector is running on a local cluster (minikube or
		// microk8s), it should be accessible through the NodePort service at the
		// `localhost:30080` endpoint. Otherwise, replace `localhost` with the
		// endpoint of your cluster. If you run the app inside k8s, then you can
		// probably connect directly to the service through dns
		driver := otlpgrpc.NewDriver(
			otlpgrpc.WithInsecure(),
			otlpgrpc.WithEndpoint(otelGRPCCollector),
			otlpgrpc.WithDialOption(grpc.WithBlock()), // useful for testing
		)
		exporter, err = otlp.NewExporter(otlpTracerCtx, driver)
		if err != nil {
			log.Fatalf("failed to create exporter: %v", err)
		}
		// Return a shutdown func the caller can use to dispose tracer connection
		shutdownFunc = func() {
			log.Printf("Shutting down otlp tracer")
			err := exporter.Shutdown(otlpTracerCtx)
			if err != nil {
				log.Fatalf("failed to stop exporter: %v", err)
			}
		}
	}

	// For the demonstration, use sdktrace.AlwaysSample sampler to sample all traces.
	// 	In a production application, use sdktrace.ProbabilitySampler with a desired probability.
	tp := sdktrace.NewTracerProvider(sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithSyncer(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(semconv.ServiceNameKey.String(serviceName))))
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return shutdownFunc
}
