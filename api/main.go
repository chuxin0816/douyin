package main

import (
	"context"
	"douyin/api/router"
	"douyin/config"
	"douyin/logger"
	"douyin/pkg/tracing"
	"douyin/rpc/client"
)

func main() {
	// 加载配置
	config.Init()

	// 初始化链路追踪
	tracing.Init(context.Background(), config.Conf.OpenTelemetryConfig.ApiName)
	defer tracing.Close()

	// 初始化日志
	logger.Init()

	// 初始化rpc客户端
	client.InitRpcClient()

	// 注册路由
	h := router.Setup(config.Conf.HertzConfig)

	h.Spin()
}
