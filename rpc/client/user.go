package client

import (
	"context"
	"douyin/config"
	"douyin/rpc/kitex_gen/user"
	"douyin/rpc/kitex_gen/user/userservice"

	"github.com/cloudwego/kitex/client"
	consul "github.com/kitex-contrib/registry-consul"
)

var userClient userservice.Client

func initUserClient() {
	// 服务发现
	r, err := consul.NewConsulResolver(config.Conf.ConsulConfig.ConsulAddr)
	if err != nil {
		panic(err)
	}

	userClient, err = userservice.NewClient(config.Conf.ConsulConfig.UserServiceName, client.WithResolver(r))
	if err != nil {
		panic(err)
	}
}

func UserInfo(authorID, userID int64) (*user.UserInfoResponse, error) {
	return userClient.UserInfo(context.Background(), &user.UserInfoRequest{})
}
