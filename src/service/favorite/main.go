package main

import (
	"douyin/src/client"
	"douyin/src/common/kafka"
	"douyin/src/common/mtl"
	"douyin/src/common/serversuite"
	"douyin/src/common/snowflake"
	"douyin/src/config"
	"douyin/src/dal"
	"douyin/src/kitex_gen/favorite/favoriteservice"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/server"
)

func main() {
	config.Init()
	go watchConfig()
	mtl.InitTracing(config.Conf.OpenTelemetryConfig.FavoriteName)
	mtl.InitLog()
	snowflake.Init()
	dal.Init()
	defer dal.Close()
	kafka.Init()
	client.Init()

	opts := server.WithSuite(serversuite.CommonServerSuite{
		RegistryAddr: config.Conf.ConsulConfig.ConsulAddr,
		ServiceAddr:  config.Conf.ConsulConfig.FavoriteAddr,
		ServiceName:  config.Conf.OpenTelemetryConfig.FavoriteName,
	})

	svr := favoriteservice.NewServer(new(FavoriteServiceImpl), opts)

	if err := svr.Run(); err != nil {
		klog.Fatal("run server failed: ", err)
	}
}

func watchConfig() {
	for {
		select {
		case <-config.NoticeOpenTelemetry:
			mtl.InitTracing(config.Conf.OpenTelemetryConfig.FavoriteName)

		case <-config.NoticeLog:
			mtl.InitLog()

		case <-config.NoticeSnowflake:
			snowflake.Init()

		case <-config.NoticeMySQL:
			dal.InitMySQL()

		case <-config.NoticeRedis:
			dal.InitRedis()

		case <-config.NoticeKafka:
			kafka.Init()
		}
	}
}
