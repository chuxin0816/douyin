package controller

import (
	"douyin/rpc/kitex_gen/feed"
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
