package client

import (
	"context"

	"douyin/src/config"
	"douyin/src/rpc/kitex_gen/relation"
	"douyin/src/rpc/kitex_gen/relation/relationservice"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"
	consul "github.com/kitex-contrib/registry-consul"
)

var relationClient relationservice.Client

func initRelationClient() {
	// 服务发现
	r, err := consul.NewConsulResolver(config.Conf.ConsulConfig.ConsulAddr)
	if err != nil {
		panic(err)
	}

	relationClient, err = relationservice.NewClient(config.Conf.OpenTelemetryConfig.RelationName,
		client.WithResolver(r),
		client.WithSuite(tracing.NewClientSuite()),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: config.Conf.OpenTelemetryConfig.ApiName}),
	)
	if err != nil {
		panic(err)
	}
}

func RelationAction(ctx context.Context, userID, toUserID int64, actionType int64) (*relation.RelationActionResponse, error) {
	return relationClient.RelationAction(ctx, &relation.RelationActionRequest{
		UserId:     userID,
		ToUserId:   toUserID,
		ActionType: actionType,
	})
}

func FollowList(ctx context.Context, userID *int64, toUserID int64) (*relation.RelationFollowListResponse, error) {
	return relationClient.RelationFollowList(ctx, &relation.RelationFollowListRequest{
		UserId:   userID,
		ToUserId: toUserID,
	})
}

func FollowerList(ctx context.Context, userID *int64, toUserID int64) (*relation.RelationFollowerListResponse, error) {
	return relationClient.RelationFollowerList(ctx, &relation.RelationFollowerListRequest{
		UserId:   userID,
		ToUserId: toUserID,
	})
}

func FriendList(ctx context.Context, userID *int64, toUserID int64) (*relation.RelationFriendListResponse, error) {
	return relationClient.RelationFriendList(ctx, &relation.RelationFriendListRequest{
		UserId:   userID,
		ToUserId: toUserID,
	})
}
