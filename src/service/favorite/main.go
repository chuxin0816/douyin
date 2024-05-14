package main

import (
	"log"
	"net"

	"douyin/src/config"
	"douyin/src/dal"
	"douyin/src/logger"
	"douyin/src/pkg/kafka"
	"douyin/src/pkg/snowflake"
	"douyin/src/pkg/tracing"
	favorite "douyin/src/rpc/kitex_gen/favorite/favoriteservice"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	kitexTracing "github.com/kitex-contrib/obs-opentelemetry/tracing"
	consul "github.com/kitex-contrib/registry-consul"
)

func main() {
	config.Init()
	go watchConfig()
	tracing.Init(config.Conf.OpenTelemetryConfig.FavoriteName)
	defer tracing.Close()
	logger.Init()
	snowflake.Init()
	dal.Init()
	defer dal.Close()
	kafka.Init()

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
		server.WithRegistry(r),
		server.WithSuite(kitexTracing.NewServerSuite()),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: config.Conf.OpenTelemetryConfig.FavoriteName}),
		server.WithMuxTransport(),
	)
	err = svr.Run()
	if err != nil {
		log.Println(err.Error())
	}
}

func watchConfig() {
	for {
		select {
		case <-config.NoticeOpenTelemetry:
			tracing.Init(config.Conf.OpenTelemetryConfig.FavoriteName)

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
