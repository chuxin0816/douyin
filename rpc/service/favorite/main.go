package main

import (
	"douyin/config"
	"douyin/dal"
	"douyin/logger"
	"douyin/pkg/kafka"
	favorite "douyin/rpc/kitex_gen/favorite/favoriteservice"
	"log"
	"net"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	consul "github.com/kitex-contrib/registry-consul"
	"github.com/cloudwego/kitex/pkg/klog"
)

func main() {
	config.Init()
	logger.Init()
	kafka.Init()
	dal.Init()
	defer dal.Close()

	addr, err := net.ResolveTCPAddr("tcp", config.Conf.ConsulConfig.FavoriteAddr)
	if err != nil {
		klog.Fatal("resolve tcp addr failed: ", err)
	}

	// 服务注册
	r, err := consul.NewConsulRegister(config.Conf.ConsulConfig.ConsulAddr)
	if err != nil {
		klog.Fatal("new consul register failed: ", err)
	}

	svr := favorite.NewServer(new(FavoriteServiceImpl),
		server.WithServiceAddr(addr),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: config.Conf.FavoriteServiceName}),
		server.WithRegistry(r),
	)
	err = svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
