package main

import (
	"douyin/config"
	"douyin/dal"
	"douyin/logger"
	"douyin/pkg/oss"
	"douyin/pkg/snowflake"
	"douyin/router"
)

func main() {
	// 加载配置
	config.Init()

	// 初始化日志
	logger.Init(config.Conf.LogConfig)

	// 初始化database
	dal.Init(config.Conf.DatabaseConfig)
	defer dal.Close()

	// 初始化雪花算法
	snowflake.Init(config.Conf.SnowflakeConfig)

	// 初始化oss
	oss.Init(config.Conf.OssConfig)

	// 注册路由
	h := router.Setup(config.Conf.HertzConfig)
	// 启动服务
	h.Spin()
}
