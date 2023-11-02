package main

import (
	"douyin/config"
	"douyin/dao/mysql"
	"douyin/dao/redis"
	"douyin/logger"
	"douyin/pkg/oss"
	"douyin/pkg/snowflake"
	"douyin/router"
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
	// 初始化mysql
	if err := mysql.Init(config.Conf.MysqlConfig); err != nil {
		hlog.Error("mysql init failed, err: ", err)
		return
	}
	// 初始化redis
	if err := redis.Init(config.Conf.RedisConfig); err != nil {
		hlog.Error("redis init failed, err: ", err)
		return
	}
	defer redis.Close()
	// 初始化雪花算法
	if err := snowflake.Init(config.Conf.StartTime, config.Conf.MachineID); err != nil {
		hlog.Error("snowflake init failed, err: ", err)
		return
	}
	// 初始化oss
	if err := oss.Init(config.Conf.OssConfig); err != nil {
		hlog.Error("oss init failed, err: ", err)
		return
	}
	// 注册路由
	h := router.Setup(config.Conf.HertzConfig)
	// 启动服务
	h.Spin()
}
