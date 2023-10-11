package controller

import (
	"context"
	"douyin/models"
	"douyin/response"
	"douyin/service"
	"fmt"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// Feed 不限制登录状态，返回按投稿时间倒序的视频列表，视频数由服务端控制，单次最多30个
func Feed(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &models.FeedRequest{LatestTime: strconv.FormatInt(time.Now().Unix(), 10)}
	ctx.Bind(req)

	// 业务逻辑处理
	resp, err := service.Feed(req)
	if err != nil {
		ctx.JSON(consts.StatusInternalServerError, nil)
		hlog.Error("controller.Feed: 业务逻辑处理失败", err)
		return
	}

	// 返回结果
	for _, video := range resp.VideoList {
		fmt.Println(*video)
	}
	response.Success(ctx, response.FeedResponse{
		Response:  &response.Response{StatusCode: response.CodeSuccess, StatusMsg: resp.StatusCode.Msg()},
		NextTime:  resp.NextTime,
		VideoList: DemoVideos,
	})
}

var DemoVideos = []*response.VideoResponse{
	{
		ID:            1,
		Author:        &DemoUser,
		PlayURL:       "https://www.w3schools.com/html/movie.mp4",
		CoverURL:      "https://cdn.pixabay.com/photo/2016/03/27/18/10/bear-1283347_1280.jpg",
		FavoriteCount: 0,
		CommentCount:  0,
		IsFavorite:    false,
	},
}

var DemoUser = response.UserResponse{
	ID:            1,
	Name:          "TestUser",
	FollowCount:   0,
	FollowerCount: 0,
	IsFollow:      false,
}