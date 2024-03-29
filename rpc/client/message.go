package client

import (
	"context"
	"douyin/config"
	"douyin/rpc/kitex_gen/message"
	"douyin/rpc/kitex_gen/message/messageservice"

	"github.com/cloudwego/kitex/client"
	consul "github.com/kitex-contrib/registry-consul"
)

var messageClient messageservice.Client

func initMessageClient() {
	// 服务发现
	r, err := consul.NewConsulResolver(config.Conf.ConsulConfig.ConsulAddr)
	if err != nil {
		panic(err)
	}

	messageClient, err = messageservice.NewClient(config.Conf.ConsulConfig.MessageServiceName, client.WithResolver(r))
	if err != nil {
		panic(err)
	}
}

func MessageAction(ctx context.Context, userID, toUserID, actionType int64, content string) (*message.MessageActionResponse, error) {
	return messageClient.MessageAction(ctx, &message.MessageActionRequest{
		UserId:     userID,
		ToUserId:   toUserID,
		ActionType: actionType,
		Content:    content,
	})
}

func MessageChat(ctx context.Context, userID, toUserID, lastTime int64) (*message.MessageChatResponse, error) {
	return messageClient.MessageChat(ctx, &message.MessageChatRequest{
		UserId:   userID,
		ToUserId: toUserID,
		LastTime: lastTime,
	})
}
