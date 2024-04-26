package main

import (
	"context"
	"net"

	"douyin/config"
	"douyin/dal"
	"douyin/logger"
	"douyin/pkg/kafka"
	"douyin/pkg/snowflake"
	"douyin/pkg/tracing"
	feed "douyin/rpc/kitex_gen/feed/feedservice"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	kitexTracing "github.com/kitex-contrib/obs-opentelemetry/tracing"
	consul "github.com/kitex-contrib/registry-consul"
)

func main() {
	config.Init()
	tracing.Init(context.Background(), config.Conf.OpenTelemetryConfig.FeedName)
	defer tracing.Close()
	logger.Init()
	snowflake.Init()
	dal.Init()
	defer dal.Close()
	kafka.Init()

	addr, err := net.ResolveTCPAddr("tcp", config.Conf.ConsulConfig.FeedAddr)
	if err != nil {
		klog.Fatal("resolve tcp addr failed: ", err)
	}

	// 服务注册
	r, err := consul.NewConsulRegister(config.Conf.ConsulConfig.ConsulAddr)
	if err != nil {
		klog.Fatal("new consul register failed: ", err)
	}

	svr := feed.NewServer(new(FeedServiceImpl),
		server.WithServiceAddr(addr),
		server.WithRegistry(r),
		server.WithSuite(kitexTracing.NewServerSuite()),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: config.Conf.OpenTelemetryConfig.FeedName}),
	)

	if err = svr.Run(); err != nil {
		klog.Fatal("run server failed: ", err)
	}
}
