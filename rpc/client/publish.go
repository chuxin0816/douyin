package client

import (
	"context"
	"douyin/config"
	"douyin/rpc/kitex_gen/publish"
	"douyin/rpc/kitex_gen/publish/publishservice"

	"github.com/cloudwego/kitex/client"
	consul "github.com/kitex-contrib/registry-consul"
)

var publishClient publishservice.Client

func initPublishClient() {
	// 服务发现
	r, err := consul.NewConsulResolver(config.Conf.ConsulConfig.ConsulAddr)
	if err != nil {
		panic(err)
	}

	publishClient, err = publishservice.NewClient(config.Conf.ConsulConfig.PublishServiceName, client.WithResolver(r))
	if err != nil {
		panic(err)
	}
}

func PublishAction(ctx context.Context, userID int64, data []byte, title string) (*publish.PublishActionResponse, error) {
	return publishClient.PublishAction(ctx, &publish.PublishActionRequest{
		UserId: userID,
		Data:   data,
		Title:  title,
	})
}

func PublishList(ctx context.Context, userID *int64, toUserID int64) (*publish.PublishListResponse, error) {
	return publishClient.PublishList(ctx, &publish.PublishListRequest{
		UserId:   userID,
		ToUserId: toUserID,
	})
}
