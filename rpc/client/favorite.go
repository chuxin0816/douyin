package client

import (
	"context"
	"douyin/config"
	"douyin/rpc/kitex_gen/favorite"
	"douyin/rpc/kitex_gen/favorite/favoriteservice"

	"github.com/cloudwego/kitex/client"
	consul "github.com/kitex-contrib/registry-consul"
)

var favoriteClient favoriteservice.Client

func initFavoriteClient() {
	// 服务发现
	r, err := consul.NewConsulResolver(config.Conf.ConsulConfig.ConsulAddr)
	if err != nil {
		panic(err)
	}

	favoriteClient, err = favoriteservice.NewClient(config.Conf.ConsulConfig.FavoriteServiceName, client.WithResolver(r))
	if err != nil {
		panic(err)
	}
}

func FavoriteAction(userID, videoID int64, actionType int64) (*favorite.FavoriteActionResponse, error) {
	return favoriteClient.FavoriteAction(context.Background(), &favorite.FavoriteActionRequest{
		UserId:     userID,
		VideoId:    videoID,
		ActionType: actionType,
	})
}

func FavoriteList(userID, toUserID int64) (*favorite.FavoriteListResponse, error) {
	return favoriteClient.FavoriteList(context.Background(), &favorite.FavoriteListRequest{
		UserId:   userID,
		ToUserId: toUserID,
	})
}
