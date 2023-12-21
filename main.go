package main

import (
	"douyin/config"
	"douyin/logger"
	"douyin/router"
	"douyin/rpc/client"
)

func main() {
	// 加载配置
	config.Init()

	// 初始化日志
	logger.Init()

	// 初始化rpc客户端
	client.InitRpcClient()

	// 注册路由
	h := router.Setup(config.Conf.HertzConfig)

	h.Spin()
}
