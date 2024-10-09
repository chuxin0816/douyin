package main

import (
	"douyin/src/client"
	"douyin/src/common/kafka"
	"douyin/src/common/mtl"
	"douyin/src/common/serversuite"
	"douyin/src/common/snowflake"
	"douyin/src/config"
	"douyin/src/dal"
	"douyin/src/kitex_gen/comment/commentservice"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/server"
)

func main() {
	config.Init()
	go watchConfig()
	mtl.Init(config.Conf.OpenTelemetryConfig.CommentName)
	defer mtl.Close()
	snowflake.Init()
	dal.Init()
	defer dal.Close()
	kafka.Init()
	client.Init()

	opts := server.WithSuite(serversuite.CommonServerSuite{
		RegistryAddr: config.Conf.ConsulConfig.ConsulAddr,
		ServiceAddr:  config.Conf.ConsulConfig.CommentAddr,
		ServiceName:  config.Conf.OpenTelemetryConfig.CommentName,
	})

	svr := commentservice.NewServer(new(CommentServiceImpl), opts)

	if err := svr.Run(); err != nil {
		klog.Fatal("run server failed: ", err)
	}
}

func watchConfig() {
	for {
		select {
		case <-config.NoticeOpenTelemetry:
			mtl.InitTracing(config.Conf.OpenTelemetryConfig.CommentName)

		case <-config.NoticeLog:
			mtl.InitLog()

		case <-config.NoticeSnowflake:
			snowflake.Init()

		case <-config.NoticePostgreSQL:
			dal.InitPostgreSQL()

		case <-config.NoticeRedis:
			dal.InitRedis()

		case <-config.NoticeKafka:
			kafka.Init()
		}
	}
}
