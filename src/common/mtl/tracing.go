package mtl

import (
	"context"
	"douyin/src/config"

	"github.com/kitex-contrib/obs-opentelemetry/provider"
)

var p provider.OtelProvider

func InitTracing(serviceName string) {
	p = provider.NewOpenTelemetryProvider(
		provider.WithServiceName(serviceName),
		provider.WithExportEndpoint(config.Conf.JaegerAddr),
		provider.WithInsecure(),
		provider.WithEnableMetrics(false),
	)
}

func ShutdownTracing() {
	p.Shutdown(context.Background()) //nolint:errcheck
}
