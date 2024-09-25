package client

import (
	"douyin/src/common/clientsuite"
	"douyin/src/config"
	"douyin/src/kitex_gen/favorite/favoriteservice"

	"github.com/cloudwego/kitex/client"
)

func initFavoriteClient() {
	opts := []client.Option{
		client.WithSuite(clientsuite.CommonClientSuite{
			RegistryAddr: config.Conf.ConsulConfig.ConsulAddr,
			ServiceName:  config.Conf.OpenTelemetryConfig.FavoriteName,
		}),
	}

	FavoriteClient, err = favoriteservice.NewClient(config.Conf.OpenTelemetryConfig.FavoriteName, opts...)
	if err != nil {
		panic(err)
	}
}
