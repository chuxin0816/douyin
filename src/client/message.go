package client

import (
	"douyin/src/common/clientsuite"
	"douyin/src/config"
	"douyin/src/kitex_gen/message/messageservice"

	"github.com/cloudwego/kitex/client"
)

func initMessageClient() {
	opts := []client.Option{
		client.WithSuite(clientsuite.CommonClientSuite{
			RegistryAddr: config.Conf.ConsulConfig.ConsulAddr,
			ServiceName:  config.Conf.OpenTelemetryConfig.MessageName,
		}),
	}

	MessageClient, err = messageservice.NewClient(config.Conf.OpenTelemetryConfig.MessageName, opts...)
	if err != nil {
		panic(err)
	}
}
