package client

import (
	"context"
	"douyin/config"
	"douyin/rpc/kitex_gen/comment"
	"douyin/rpc/kitex_gen/comment/commentservice"

	"github.com/cloudwego/kitex/client"
	consul "github.com/kitex-contrib/registry-consul"
)

var commentClient commentservice.Client

func initCommentClient() {
	// 服务发现
	r, err := consul.NewConsulResolver(config.Conf.ConsulConfig.ConsulAddr)
	if err != nil {
		panic(err)
	}

	commentClient, err = commentservice.NewClient(config.Conf.ConsulConfig.CommentServiceName, client.WithResolver(r))
	if err != nil {
		panic(err)
	}
}

func CommentAction(ctx context.Context, userID, videoID, actionType int64, commentID *int64, commentText *string) (*comment.CommentActionResponse, error) {
	return commentClient.CommentAction(ctx, &comment.CommentActionRequest{
		UserId:      userID,
		VideoId:     videoID,
		ActionType:  actionType,
		CommentText: commentText,
		CommentId:   commentID,
	})
}

func CommentList(ctx context.Context, userID *int64, videoID int64) (*comment.CommentListResponse, error) {
	return commentClient.CommentList(ctx, &comment.CommentListRequest{
		UserId:  userID,
		VideoId: videoID,
	})
}
