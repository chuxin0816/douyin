package main

import (
	"net"

	"douyin/src/config"
	"douyin/src/dal"
	message "douyin/src/kitex_gen/message/messageservice"
	"douyin/src/logger"
	"douyin/src/pkg/kafka"
	"douyin/src/pkg/snowflake"
	"douyin/src/pkg/tracing"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	kitexTracing "github.com/kitex-contrib/obs-opentelemetry/tracing"
	consul "github.com/kitex-contrib/registry-consul"
)

func main() {
	config.Init()
	go watchConfig()
	tracing.Init(config.Conf.OpenTelemetryConfig.MessageName)
	defer tracing.Close()
	logger.Init()
	snowflake.Init()
	dal.Init()
	defer dal.Close()
	kafka.Init()

	addr, err := net.ResolveTCPAddr("tcp", config.Conf.ConsulConfig.MessageAddr)
	if err != nil {
		klog.Fatal("resolve tcp addr failed: ", err)
	}

	// 服务注册
	r, err := consul.NewConsulRegister(config.Conf.ConsulConfig.ConsulAddr)
	if err != nil {
		klog.Fatal("new consul register failed: ", err)
	}

	svr := message.NewServer(new(MessageServiceImpl),
		server.WithServiceAddr(addr),
		server.WithRegistry(r),
		server.WithSuite(kitexTracing.NewServerSuite()),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: config.Conf.OpenTelemetryConfig.MessageName}),
		server.WithMuxTransport(),
	)

	if err = svr.Run(); err != nil {
		klog.Fatal("run server failed: ", err)
	}
}

func watchConfig() {
	for {
		select {
		case <-config.NoticeOpenTelemetry:
			tracing.Init(config.Conf.OpenTelemetryConfig.MessageName)

		case <-config.NoticeLog:
			logger.Init()

		case <-config.NoticeSnowflake:
			snowflake.Init()

		case <-config.NoticeRedis:
			dal.InitRedis()

		case <-config.NoticeMongo:
			dal.InitMongo()

		case <-config.NoticeKafka:
			kafka.Init()
		}
	}
}
