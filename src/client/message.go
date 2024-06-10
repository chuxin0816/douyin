package client

import (
	"douyin/src/config"
	"douyin/src/kitex_gen/message/messageservice"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	consul "github.com/kitex-contrib/registry-consul"
)

func initMessageClient() {
	// 服务发现
	r, err := consul.NewConsulResolver(config.Conf.ConsulConfig.ConsulAddr)
	if err != nil {
		panic(err)
	}

	MessageClient, err = messageservice.NewClient(
		config.Conf.OpenTelemetryConfig.MessageName,
		client.WithResolver(r),
		client.WithSuite(tracing.NewClientSuite()),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: config.Conf.OpenTelemetryConfig.MessageName}),
		client.WithMuxConnection(2),
	)
	if err != nil {
		panic(err)
	}
}
