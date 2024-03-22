package otel

import (
	"context"
	"douyin/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

var tp *trace.TracerProvider

func Init(ctx context.Context, serviceName string) {
	exporter, err := otlptracehttp.New(ctx,
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
}

func Close() {
	if err := tp.Shutdown(context.Background()); err != nil {
		panic(err)
	}
}
