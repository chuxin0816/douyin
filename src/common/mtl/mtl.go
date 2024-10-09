package mtl

import "douyin/src/config"

func Init(serviceName string) {
	// 初始化监控指标
	InitMetric(serviceName, config.Conf.OpenTelemetryConfig.MetricAddr, config.Conf.ConsulConfig.ConsulAddr)

	// 初始化链路追踪
	InitTracing(serviceName)

	// 初始化日志
	InitLog()
}

func Close() {
	DeregisterMetric()
	ShutdownTracing()
}
