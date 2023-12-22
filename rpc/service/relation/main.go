package main

import (
	"douyin/config"
	"douyin/dal"
	"douyin/logger"
	relation "douyin/rpc/kitex_gen/relation/relationservice"
	"net"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	consul "github.com/kitex-contrib/registry-consul"
	"github.com/u2takey/go-utils/klog"
)

func main() {
	config.Init()
	logger.Init()
	dal.Init()

	addr, err := net.ResolveTCPAddr("tcp", config.Conf.ConsulConfig.RelationAddr)
	if err != nil {
		klog.Fatal("resolve tcp addr failed: ", err)
	}

	// 服务注册
	r, err := consul.NewConsulRegister(config.Conf.ConsulConfig.ConsulAddr)
	if err != nil {
		klog.Fatal("new consul register failed: ", err)
	}

	svr := relation.NewServer(new(RelationServiceImpl),
		server.WithServiceAddr(addr),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: config.Conf.RelationServiceName}),
		server.WithRegistry(r),
	)

	if err = svr.Run(); err != nil {
		klog.Fatal("run server failed: ", err)
	}

}
