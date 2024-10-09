package main

import (
	"douyin/src/client"
	"douyin/src/common/jwt"
	"douyin/src/common/mtl"
	"douyin/src/config"
	"douyin/src/dal"
	"douyin/src/service/api/router"
)

func main() {
	// 加载配置
	config.Init()
	go watchConfig()

	// 初始化jwt
	jwt.Init()

	// 初始化监控指标,链路追踪,日志
	mtl.Init()
	defer mtl.Close()

	// 初始化RPC客户端
	client.Init()

	// 初始化Redis
	dal.InitRedis()
	defer dal.Close()

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

		case <-config.NoticeRedis:
			dal.InitRedis()
		}
	}
}
