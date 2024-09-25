package client

import (
	"douyin/src/common/clientsuite"
	"douyin/src/config"
	"douyin/src/kitex_gen/user/userservice"

	"github.com/cloudwego/kitex/client"
)

func initUserClient() {
	opts := []client.Option{
		client.WithSuite(clientsuite.CommonClientSuite{
			RegistryAddr: config.Conf.ConsulConfig.ConsulAddr,
			ServiceName:  config.Conf.OpenTelemetryConfig.UserName,
		}),
	}

	UserClient, err = userservice.NewClient(config.Conf.OpenTelemetryConfig.UserName, opts...)
	if err != nil {
		panic(err)
	}
}
