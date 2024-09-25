package client

import (
	"douyin/src/common/clientsuite"
	"douyin/src/config"
	"douyin/src/kitex_gen/relation/relationservice"

	"github.com/cloudwego/kitex/client"
)

func initRelationClient() {
	opts := []client.Option{
		client.WithSuite(clientsuite.CommonClientSuite{
			RegistryAddr: config.Conf.ConsulConfig.ConsulAddr,
			ServiceName:  config.Conf.OpenTelemetryConfig.RelationName,
		}),
	}

	RelationClient, err = relationservice.NewClient(config.Conf.OpenTelemetryConfig.RelationName, opts...)
	if err != nil {
		panic(err)
	}
}
