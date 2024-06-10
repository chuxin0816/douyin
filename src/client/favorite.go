package client

import (
	"douyin/src/config"
	"douyin/src/kitex_gen/favorite/favoriteservice"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	consul "github.com/kitex-contrib/registry-consul"
)

func initFavoriteClient() {
	// 服务发现
	r, err := consul.NewConsulResolver(config.Conf.ConsulConfig.ConsulAddr)
	if err != nil {
		panic(err)
	}

	FavoriteClient, err = favoriteservice.NewClient(
		config.Conf.OpenTelemetryConfig.FavoriteName,
		client.WithResolver(r),
		client.WithSuite(tracing.NewClientSuite()),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: config.Conf.OpenTelemetryConfig.FavoriteName}),
		client.WithMuxConnection(2),
	)
	if err != nil {
		panic(err)
	}
}
