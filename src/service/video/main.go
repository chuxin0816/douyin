package main

import (
	"douyin/src/client"
	"douyin/src/common/kafka"
	"douyin/src/common/mtl"
	"douyin/src/common/oss"
	"douyin/src/common/serversuite"
	"douyin/src/common/snowflake"
	"douyin/src/config"
	"douyin/src/dal"
	"douyin/src/kitex_gen/video/videoservice"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/server"
)

func main() {
	config.Init()
	go watchConfig()
	mtl.InitTracing(config.Conf.OpenTelemetryConfig.VideoName)
	mtl.InitLog()
	snowflake.Init()
	dal.Init()
	defer dal.Close()
	kafka.Init()
	client.Init()

	opts := server.WithSuite(serversuite.CommonServerSuite{
		RegistryAddr: config.Conf.ConsulConfig.ConsulAddr,
		ServiceAddr:  config.Conf.ConsulConfig.VideoAddr,
		ServiceName:  config.Conf.OpenTelemetryConfig.VideoName,
	})

	svr := videoservice.NewServer(new(VideoServiceImpl), opts)

	if err := svr.Run(); err != nil {
		klog.Fatal("run server failed: ", err)
	}
}

func watchConfig() {
	for {
		select {
		case <-config.NoticeOpenTelemetry:
			mtl.InitTracing(config.Conf.OpenTelemetryConfig.VideoName)

		case <-config.NoticeLog:
			mtl.InitLog()

		case <-config.NoticeSnowflake:
			snowflake.Init()

		case <-config.NoticeOss:
			oss.Init()

		case <-config.NoticeMySQL:
			dal.InitMySQL()

		case <-config.NoticeRedis:
			dal.InitRedis()

		case <-config.NoticeKafka:
			kafka.Init()
		}
	}
}
