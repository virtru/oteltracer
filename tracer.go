package oteltracer

import (
	"context"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"

	"google.golang.org/grpc"
)

var otelGRPCCollector = os.Getenv("OTLP_COLLECTOR_ENDPOINT") //ex. localhost:4345

func InitTracer(serviceName string) (func(), error) {
	var shutdownFunc func()
	var err error
	if otelGRPCCollector == "" {
		exp, err := stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
		if err != nil {
			return shutdownFunc, err
		}
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithSyncer(exp),
		)
		otel.SetTracerProvider(tp)
		shutdownFunc = func() {
			log.Printf("Shutting down stdout tracer")
		}
		return shutdownFunc, nil
	}

	log.Printf("Starting OpenTelemetry tracer: otlp, configured with endpoint: %s", otelGRPCCollector)
	//Create a bounded context so we don't block indefinitely waiting to connect
	otlpTracerCtx, otlpContextCancel := context.WithTimeout(context.Background(), time.Second*10)

	// If the OpenTelemetry Collector is running on a local cluster (minikube or
	// microk8s), it should be accessible through the NodePort service at the
	// `localhost:30080` endpoint. Otherwise, replace `localhost` with the
	// endpoint of your cluster. If you run the app inside k8s, then you can
	// probably connect directly to the service through dns
	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(otelGRPCCollector),
		otlptracegrpc.WithDialOption(grpc.WithBlock()), // Block calling app unless and until we make a connection, unless our context times out first
	)
	if err != nil {
		log.Printf("ERROR: failed to create otlp driver: %v", err)
		otlpContextCancel()
		return shutdownFunc, err
	}
	exporter, err := otlptrace.New(otlpTracerCtx, client)
	if err != nil {
		log.Printf("ERROR: failed to create exporter: %v", err)
		otlpContextCancel()
		return shutdownFunc, err
	}
	// Return a shutdown func the caller can use to dispose tracer connection
	shutdownFunc = func() {
		log.Printf("Shutting down otlp tracer")
		err := exporter.Shutdown(otlpTracerCtx)
		if err != nil {
			log.Printf("ERROR: failed to stop exporter: %v", err)
		}
		//Now that the exporter is shutdown, cancel our background context
		otlpContextCancel()

	}

	// For the demonstration, use sdktrace.AlwaysSample sampler to sample all traces.
	// 	In a production application, use sdktrace.ProbabilitySampler with a desired probability.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSyncer(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceNameKey.String(serviceName))))
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return shutdownFunc, nil
}
