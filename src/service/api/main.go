package main

import (
	"douyin/src/config"
	"douyin/src/logger"
	"douyin/src/pkg/jwt"
	"douyin/src/pkg/tracing"
	"douyin/src/service/api/router"
)

func main() {
	// 加载配置
	config.Init()
	go watchConfig()

	// 初始化jwt
	jwt.Init()

	// 初始化链路追踪
	tracing.Init(config.Conf.OpenTelemetryConfig.ApiName)
	defer tracing.Close()

	// 初始化日志
	logger.Init()

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
			tracing.Init(config.Conf.OpenTelemetryConfig.ApiName)

		case <-config.NoticeLog:
			logger.Init()
		}
	}
}
