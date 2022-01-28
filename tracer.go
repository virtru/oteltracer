package oteltracer

import (
	"context"
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"

	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"

	"google.golang.org/grpc"
)

var otelGRPCCollector = os.Getenv("OTLP_COLLECTOR_ENDPOINT") //ex. localhost:4345

func InitTracer(serviceName string) func() {
	var shutdownFunc func()
	var err error

	log.Printf("Starting OpenTelemetry tracer: otlp, configured with endpoint: %s", otelGRPCCollector)
	otlpTracerCtx := context.Background()

	// If the OpenTelemetry Collector is running on a local cluster (minikube or
	// microk8s), it should be accessible through the NodePort service at the
	// `localhost:30080` endpoint. Otherwise, replace `localhost` with the
	// endpoint of your cluster. If you run the app inside k8s, then you can
	// probably connect directly to the service through dns
	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(otelGRPCCollector),
		otlptracegrpc.WithDialOption(grpc.WithBlock()), // useful for testing
	)
	if err != nil {
		log.Fatalf("failed to create otlp driver: %v", err)
	}
	exporter, err := otlptrace.New(otlpTracerCtx, client)
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

	// For the demonstration, use sdktrace.AlwaysSample sampler to sample all traces.
	// 	In a production application, use sdktrace.ProbabilitySampler with a desired probability.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSyncer(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceNameKey.String(serviceName))))
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return shutdownFunc
}
