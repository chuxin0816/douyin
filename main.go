package main

import (
	"douyin/config"
	"douyin/controller"
	"douyin/dal"
	"douyin/logger"
	"douyin/pkg/oss"
	"douyin/pkg/snowflake"
	"fmt"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func main() {
	// 加载配置
	if err := config.Init(); err != nil {
		fmt.Printf("config init failed, err:%v\n", err)
		return
	}
	// 初始化日志
	if err := logger.Init(config.Conf.LogConfig); err != nil {
		fmt.Printf("logger init failed, err:%v\n", err)
		return
	}
	// 初始化database
	if err := dal.Init(config.Conf.DatabaseConfig); err != nil {
		hlog.Error("dao init failed, err: ", err)
		return
	}
	defer dal.Close()
	// 初始化雪花算法
	if err := snowflake.Init(config.Conf.SnowflakeConfig); err != nil {
		hlog.Error("snowflake init failed, err: ", err)
		return
	}
	// 初始化oss
	if err := oss.Init(config.Conf.OssConfig); err != nil {
		hlog.Error("oss init failed, err: ", err)
		return
	}
	// 注册路由
	h := controller.Setup(config.Conf.HertzConfig)
	// 启动服务
	h.Spin()
}
