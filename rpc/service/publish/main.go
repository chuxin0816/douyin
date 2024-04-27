package main

import (
	"net"

	"douyin/config"
	"douyin/dal"
	"douyin/logger"
	"douyin/pkg/kafka"
	"douyin/pkg/oss"
	"douyin/pkg/snowflake"
	"douyin/pkg/tracing"
	publish "douyin/rpc/kitex_gen/publish/publishservice"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	kitexTracing "github.com/kitex-contrib/obs-opentelemetry/tracing"
	consul "github.com/kitex-contrib/registry-consul"
)

func main() {
	config.Init()
	go watchConfig()
	tracing.Init(config.Conf.OpenTelemetryConfig.PublishName)
	defer tracing.Close()
	logger.Init()
	snowflake.Init()
	oss.Init()
	dal.Init()
	defer dal.Close()
	kafka.Init()

	addr, err := net.ResolveTCPAddr("tcp", config.Conf.ConsulConfig.PublishAddr)
	if err != nil {
		klog.Fatal("resolve tcp addr failed: ", err)
	}

	// 服务注册
	r, err := consul.NewConsulRegister(config.Conf.ConsulConfig.ConsulAddr)
	if err != nil {
		klog.Fatal("new consul register failed: ", err)
	}

	svr := publish.NewServer(new(PublishServiceImpl),
		server.WithServiceAddr(addr),
		server.WithRegistry(r),
		server.WithSuite(kitexTracing.NewServerSuite()),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: config.Conf.OpenTelemetryConfig.PublishName}),
	)

	if err = svr.Run(); err != nil {
		klog.Fatal("run server failed: ", err)
	}
}

func watchConfig() {
	for {
		select {
		case <-config.NoticeOpenTelemetry:
			tracing.Init(config.Conf.OpenTelemetryConfig.PublishName)

		case <-config.NoticeLog:
			logger.Init()

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
