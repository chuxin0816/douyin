package tracing

import (
	"context"

	"douyin/src/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	trace2 "go.opentelemetry.io/otel/trace"
)

var (
	tp     *trace.TracerProvider
	Tracer trace2.Tracer
)

func Init(serviceName string) {
	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(config.Conf.OpenTelemetryConfig.JaegerAddr),
		otlptracehttp.WithInsecure())
	if err != nil {
		panic(err)
	}

	tp = trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	Tracer = otel.Tracer(serviceName)
}

func Close() {
	if err := tp.Shutdown(context.Background()); err != nil {
		panic(err)
	}
}
