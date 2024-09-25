package main

import (
	"douyin/src/client"
	"douyin/src/common/jwt"
	"douyin/src/common/mtl"
	"douyin/src/config"
	"douyin/src/service/api/router"
)

func main() {
	// 加载配置
	config.Init()
	go watchConfig()

	// 初始化jwt
	jwt.Init()

	// 初始化监控指标
	mtl.InitMetric(config.Conf.OpenTelemetryConfig.ApiName, config.Conf.OpenTelemetryConfig.MetricAddr, config.Conf.ConsulConfig.ConsulAddr)
	defer mtl.DeregisterMetric()

	// 初始化链路追踪
	mtl.InitTracing(config.Conf.OpenTelemetryConfig.ApiName)
	defer mtl.ShutdownTracing()

	// 初始化日志
	mtl.InitLog()

	// 初始化RPC客户端
	client.Init()

	// 注册路由
	h := router.Setup(config.Conf.HertzConfig)

	h.Spin()
}

func watchConfig() {
	for {
		select {
		case <-config.NoticeJwt:
			jwt.Init()

		case <-config.NoticeOpenTelemetry:
			mtl.InitTracing(config.Conf.OpenTelemetryConfig.ApiName)

		case <-config.NoticeLog:
			mtl.InitLog()
		}
	}
}
