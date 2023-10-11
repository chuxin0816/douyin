package main

import (
	"douyin/config"
	"douyin/dao/mysql"
	"douyin/dao/redis"
	"douyin/logger"
	"douyin/router"
	"fmt"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func main() {
	// 1.加载配置
	if err := config.Init(); err != nil {
		fmt.Printf("config init failed, err:%v\n", err)
		return
	}
	// 2.初始化日志
	if err := logger.Init(config.Conf.LogConfig); err != nil {
		fmt.Printf("logger init failed, err:%v\n", err)
		return
	}
	// 3.初始化mysql
	if err := mysql.Init(config.Conf.MysqlConfig); err != nil {
		hlog.Error("mysql init failed, err:%v", err)
		return
	}
	// 4.初始化redis
	if err := redis.Init(config.Conf.RedisConfig); err != nil {
		hlog.Error("redis init failed, err:%v", err)
		return
	}
	defer redis.RDB.Close()
	// 5.注册路由
	h := router.Setup(config.Conf.HertzConfig)
	// 6.启动服务
	h.Spin()
}
