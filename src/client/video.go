package client

import (
	"douyin/src/common/clientsuite"
	"douyin/src/config"
	"douyin/src/kitex_gen/video/videoservice"

	"github.com/cloudwego/kitex/client"
)

func initVideoClient() {
	opts := []client.Option{
		client.WithSuite(clientsuite.CommonClientSuite{
			RegistryAddr: config.Conf.ConsulConfig.ConsulAddr,
			ServiceName:  config.Conf.OpenTelemetryConfig.VideoName,
		}),
	}

	VideoClient, err = videoservice.NewClient(config.Conf.OpenTelemetryConfig.VideoName, opts...)
	if err != nil {
		panic(err)
	}
}
