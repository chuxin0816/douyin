package mtl

import (
	"context"

	"github.com/cloudwego/kitex/server"
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

	server.RegisterShutdownHook(func() {
		p.Shutdown(context.Background()) //nolint:errcheck
	})
}

func ShutdownTracing() error {
	return p.Shutdown(context.Background())
}
