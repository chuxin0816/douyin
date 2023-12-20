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

func FavoriteAction(token string, videoID int64, actionType string) (*favorite.FavoriteActionResponse, error) {
	return favoriteClient.FavoriteAction(context.Background(), &favorite.FavoriteActionRequest{
		Token:      token,
		VideoId:    videoID,
		ActionType: actionType,
	})
}

func FavoriteList(userID int64, token string) (*favorite.FavoriteListResponse, error) {
	return favoriteClient.FavoriteList(context.Background(), &favorite.FavoriteListRequest{
		UserId: userID,
		Token:  token,
	})
}
