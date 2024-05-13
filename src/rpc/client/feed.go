package client

import (
	"context"

	"douyin/src/config"
	"douyin/src/rpc/kitex_gen/feed"
	"douyin/src/rpc/kitex_gen/feed/feedservice"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	consul "github.com/kitex-contrib/registry-consul"
)

var feedClient feedservice.Client

func initFeedClient() {
	// 服务发现
	r, err := consul.NewConsulResolver(config.Conf.ConsulConfig.ConsulAddr)
	if err != nil {
		panic(err)
	}

	feedClient, err = feedservice.NewClient(config.Conf.OpenTelemetryConfig.FeedName,
		client.WithResolver(r),
		client.WithSuite(tracing.NewClientSuite()),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: config.Conf.OpenTelemetryConfig.ApiName}),
	)
	if err != nil {
		panic(err)
	}
}

func Feed(ctx context.Context, latestTime int64, userID *int64) (*feed.FeedResponse, error) {
	return feedClient.Feed(ctx, &feed.FeedRequest{
		LatestTime: latestTime,
		UserId:     userID,
	})
}
