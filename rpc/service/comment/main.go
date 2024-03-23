package main

import (
	"context"
	"douyin/config"
	"douyin/dal"
	"douyin/logger"
	"douyin/pkg/kafka"
	"douyin/pkg/snowflake"
	"douyin/pkg/trace"
	comment "douyin/rpc/kitex_gen/comment/commentservice"
	"net"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	consul "github.com/kitex-contrib/registry-consul"
)

func main() {
	config.Init()
	trace.Init(context.Background(), config.Conf.OpenTelemetryConfig.CommentName)
	defer trace.Close()
	logger.Init()
	snowflake.Init()
	kafka.Init()
	dal.Init()
	defer dal.Close()

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
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: config.Conf.CommentServiceName}),
		server.WithRegistry(r),
	)

	if err = svr.Run(); err != nil {
		klog.Fatal("run server failed: ", err)
	}
}
