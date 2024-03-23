package main

import (
	"context"
	"douyin/api/router"
	"douyin/config"
	"douyin/logger"
	"douyin/pkg/trace"
	"douyin/rpc/client"
)

func main() {
	// 加载配置
	config.Init()

	trace.Init(context.Background(), config.Conf.OpenTelemetryConfig.ApiName)
	defer trace.Close()

	// 初始化日志
	logger.Init()

	// 初始化rpc客户端
	client.InitRpcClient()

	// 注册路由
	h := router.Setup(config.Conf.HertzConfig)

	h.Spin()
}
