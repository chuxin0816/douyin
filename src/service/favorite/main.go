package main

import (
	"log"
	"net"

	"douyin/src/client"
	"douyin/src/common/kafka"
	"douyin/src/common/mtl"
	"douyin/src/common/snowflake"
	"douyin/src/config"
	"douyin/src/dal"
	"douyin/src/kitex_gen/favorite/favoriteservice"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	prometheus "github.com/kitex-contrib/monitor-prometheus"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	consul "github.com/kitex-contrib/registry-consul"
)

func main() {
	config.Init()
	go watchConfig()
	mtl.InitTracing(config.Conf.OpenTelemetryConfig.FavoriteName)
	mtl.InitLog()
	snowflake.Init()
	dal.Init()
	defer dal.Close()
	kafka.Init()
	client.Init()

	addr, err := net.ResolveTCPAddr("tcp", config.Conf.ConsulConfig.FavoriteAddr)
	if err != nil {
		klog.Fatal("resolve tcp addr failed: ", err)
	}

	// 服务注册
	r, err := consul.NewConsulRegister(config.Conf.ConsulConfig.ConsulAddr)
	if err != nil {
		klog.Fatal("new consul register failed: ", err)
	}

	svr := favoriteservice.NewServer(new(FavoriteServiceImpl),
		server.WithServiceAddr(addr),
		server.WithRegistry(r),
		server.WithSuite(tracing.NewServerSuite()),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: config.Conf.OpenTelemetryConfig.FavoriteName}),
		server.WithMuxTransport(),
		server.WithTracer(prometheus.NewServerTracer("", "", prometheus.WithDisableServer(true), prometheus.WithRegistry(mtl.Registry))),
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
			mtl.InitTracing(config.Conf.OpenTelemetryConfig.FavoriteName)

		case <-config.NoticeLog:
			mtl.InitLog()

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
