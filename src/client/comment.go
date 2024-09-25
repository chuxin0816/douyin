package client

import (
	"douyin/src/common/clientsuite"
	"douyin/src/config"
	"douyin/src/kitex_gen/comment/commentservice"

	"github.com/cloudwego/kitex/client"
)

func initCommentClient() {
	opts := []client.Option{
		client.WithSuite(clientsuite.CommonClientSuite{
			RegistryAddr: config.Conf.ConsulConfig.ConsulAddr,
			ServiceName:  config.Conf.OpenTelemetryConfig.CommentName,
		}),
	}

	CommentClient, err = commentservice.NewClient(config.Conf.OpenTelemetryConfig.CommentName, opts...)
	if err != nil {
		panic(err)
	}
}
