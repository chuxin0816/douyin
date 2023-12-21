package controller

import (
	"douyin/rpc/kitex_gen/comment"
	"douyin/rpc/kitex_gen/feed"
	"douyin/rpc/kitex_gen/message"
	"douyin/rpc/kitex_gen/user"
	"strconv"
)

func rpcUser2httpUser(rpcUser *user.User) *UserResponse {
	return &UserResponse{
		ID:              rpcUser.Id,
		Name:            rpcUser.Name,
		FollowCount:     *rpcUser.FollowCount,
		FollowerCount:   *rpcUser.FollowerCount,
		IsFollow:        rpcUser.IsFollow,
		Avatar:          *rpcUser.Avatar,
		BackgroundImage: *rpcUser.BackgroundImage,
		Signature:       *rpcUser.Signature,
		TotalFavorited:  strconv.FormatInt(*rpcUser.TotalFavorited, 10),
		WorkCount:       *rpcUser.WorkCount,
		FavoriteCount:   *rpcUser.FavoriteCount,
	}
}

func rpcVideo2httpVideo(rpcVideo *feed.Video) *VideoResponse {
	return &VideoResponse{
		ID:            rpcVideo.Id,
		Author:        rpcUser2httpUser(rpcVideo.Author),
		PlayURL:       rpcVideo.PlayUrl,
		CoverURL:      rpcVideo.CoverUrl,
		FavoriteCount: rpcVideo.FavoriteCount,
		CommentCount:  rpcVideo.CommentCount,
		IsFavorite:    rpcVideo.IsFavorite,
		Title:         rpcVideo.Title,
	}
}

func rpcComment2httpComment(rpcComment *comment.Comment) *CommentResponse {
	return &CommentResponse{
		ID:         rpcComment.Id,
		User:       rpcUser2httpUser(rpcComment.User),
		Content:    rpcComment.Content,
		CreateDate: rpcComment.CreateDate,
	}
}

func rpcMessage2httpMessage(rpcMessage *message.Message) *MessageResponse {
	return &MessageResponse{
		ID:         rpcMessage.Id,
		ToUserID:   rpcMessage.ToUserId,
		FromUserID: rpcMessage.FromUserId,
		Content:    rpcMessage.Content,
		CreateTime: rpcMessage.CreateTime,
	}
}
