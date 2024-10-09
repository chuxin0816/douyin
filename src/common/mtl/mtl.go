package mtl

import "douyin/src/config"

func Init() {
	// 初始化监控指标
	InitMetric(config.Conf.OpenTelemetryConfig.ApiName, config.Conf.OpenTelemetryConfig.MetricAddr, config.Conf.ConsulConfig.ConsulAddr)

	// 初始化链路追踪
	InitTracing(config.Conf.OpenTelemetryConfig.ApiName)

	// 初始化日志
	InitLog()
}

func Close() {
	DeregisterMetric()
	ShutdownTracing()
}
