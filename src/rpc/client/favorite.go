package client

import (
	"context"

	"douyin/src/config"
	"douyin/src/rpc/kitex_gen/favorite"
	"douyin/src/rpc/kitex_gen/favorite/favoriteservice"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	consul "github.com/kitex-contrib/registry-consul"
)

var favoriteClient favoriteservice.Client

func initFavoriteClient() {
	// 服务发现
	r, err := consul.NewConsulResolver(config.Conf.ConsulConfig.ConsulAddr)
	if err != nil {
		panic(err)
	}

	favoriteClient, err = favoriteservice.NewClient(
		config.Conf.OpenTelemetryConfig.FavoriteName,
		client.WithResolver(r),
		client.WithSuite(tracing.NewClientSuite()),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: config.Conf.OpenTelemetryConfig.ApiName}),
		client.WithMuxConnection(2),
	)
	if err != nil {
		panic(err)
	}
}

func FavoriteAction(ctx context.Context, userID, videoID int64, actionType int64) (*favorite.FavoriteActionResponse, error) {
	return favoriteClient.FavoriteAction(ctx, &favorite.FavoriteActionRequest{
		UserId:     userID,
		VideoId:    videoID,
		ActionType: actionType,
	})
}

func FavoriteList(ctx context.Context, userID *int64, toUserID int64) (*favorite.FavoriteListResponse, error) {
	return favoriteClient.FavoriteList(ctx, &favorite.FavoriteListRequest{
		UserId:   userID,
		ToUserId: toUserID,
	})
}
