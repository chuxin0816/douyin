package mtl

import (
	"context"

	"github.com/kitex-contrib/obs-opentelemetry/provider"
)

var p provider.OtelProvider

func InitTracing(serviceName string) {
	p = provider.NewOpenTelemetryProvider(
		provider.WithServiceName(serviceName),
		provider.WithExportEndpoint("localhost:4317"),
		provider.WithInsecure(),
		provider.WithEnableMetrics(false),
	)
}

func ShutdownTracing() {
	p.Shutdown(context.Background()) //nolint:errcheck
}
