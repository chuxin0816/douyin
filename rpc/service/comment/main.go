package main

import (
	"net"

	"douyin/config"
	"douyin/dal"
	"douyin/logger"
	"douyin/pkg/kafka"
	"douyin/pkg/snowflake"
	"douyin/pkg/tracing"
	comment "douyin/rpc/kitex_gen/comment/commentservice"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	kitexTracing "github.com/kitex-contrib/obs-opentelemetry/tracing"
	consul "github.com/kitex-contrib/registry-consul"
)

func main() {
	config.Init()
	go watchConfig()
	tracing.Init(config.Conf.OpenTelemetryConfig.CommentName)
	defer tracing.Close()
	logger.Init()
	snowflake.Init()
	dal.Init()
	defer dal.Close()
	kafka.Init()

	addr, err := net.ResolveTCPAddr("tcp", config.Conf.ConsulConfig.CommentAddr)
	if err != nil {
		klog.Fatal("resolve tcp addr failed: ", err)
	}

	// 服务注册
	r, err := consul.NewConsulRegister(config.Conf.ConsulConfig.ConsulAddr)
	if err != nil {
		klog.Fatal("new consul register failed: ", err)
	}

	svr := comment.NewServer(new(CommentServiceImpl),
		server.WithServiceAddr(addr),
		server.WithRegistry(r),
		server.WithSuite(kitexTracing.NewServerSuite()),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: config.Conf.OpenTelemetryConfig.CommentName}),
	)

	if err = svr.Run(); err != nil {
		klog.Fatal("run server failed: ", err)
	}
}

func watchConfig() {
	for {
		select {
		case <-config.NoticeOpenTelemetry:
			tracing.Init(config.Conf.OpenTelemetryConfig.CommentName)

		case <-config.NoticeLog:
			logger.Init()

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
