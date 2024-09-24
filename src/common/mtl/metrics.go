package mtl

import (
	"net"
	"net/http"

	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/cloudwego/kitex/server"
	consul "github.com/kitex-contrib/registry-consul"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var Registry *prometheus.Registry

func InitMetric(serviceName string, metricsPort string, registryAddr string) {
	Registry = prometheus.NewRegistry()

	// 添加 Go 编译信息
	Registry.MustRegister(collectors.NewBuildInfoCollector())
	// 添加 Go 运行时信息
	Registry.MustRegister(collectors.NewGoCollector())
	// 添加进程信息
	Registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	r, err := consul.NewConsulRegister(registryAddr)
	if err != nil {
		panic(err)
	}

	addr, err := net.ResolveTCPAddr("tcp", metricsPort)
	if err != nil {
		panic(err)
	}

	registryInfo := &registry.Info{
		ServiceName: "prometheus",
		Addr:        addr,
		Weight:      1,
		Tags:        map[string]string{"service": serviceName},
	}

	if err := r.Register(registryInfo); err != nil {
		panic(err)
	}

	server.RegisterShutdownHook(func() {
		r.Deregister(registryInfo) //nolint:errcheck
	})

	http.Handle("/metrics", promhttp.HandlerFor(Registry, promhttp.HandlerOpts{}))
	go http.ListenAndServe(metricsPort, nil) //nolint:errcheck
}
