package main

import (
	"douyin/config"
	"douyin/dal"
	"douyin/logger"
	"douyin/pkg/kafka"
	"douyin/pkg/oss"
	publish "douyin/rpc/kitex_gen/publish/publishservice"
	"net"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	consul "github.com/kitex-contrib/registry-consul"
	"github.com/cloudwego/kitex/pkg/klog"
)

func main() {
	config.Init()
	logger.Init()
	oss.Init()
	kafka.Init()
	dal.Init()
	defer dal.Close()

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
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: config.Conf.PublishServiceName}),
		server.WithRegistry(r),
	)

	if err = svr.Run(); err != nil {
		klog.Fatal("run server failed: ", err)
	}
}
