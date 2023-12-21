package client

import (
	"context"
	"douyin/config"
	"douyin/rpc/kitex_gen/comment"
	"douyin/rpc/kitex_gen/comment/commentservice"
	"douyin/rpc/kitex_gen/feed/feedservice"

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

	feedClient, err = feedservice.NewClient(config.Conf.ConsulConfig.FeedServiceName, client.WithResolver(r))
	if err != nil {
		panic(err)
	}
}

func CommentAction(userID, videoID, actionType int64, commentID *int64, commentText *string) (*comment.CommentActionResponse, error) {
	return commentClient.CommentAction(context.Background(), &comment.CommentActionRequest{
		UserId:      userID,
		VideoId:     videoID,
		ActionType:  actionType,
		CommentText: commentText,
		CommentId:   commentID,
	})
}

func CommentList(userID, videoID int64) (*comment.CommentListResponse, error) {
	return commentClient.CommentList(context.Background(), &comment.CommentListRequest{
		UserId:  userID,
		VideoId: videoID,
	})
}
