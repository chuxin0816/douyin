package controller

import (
	"context"
	"douyin/models"
	"douyin/response"
	"douyin/service"
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
	response.Success(ctx, response.FeedResponse{
		Response:  &response.Response{StatusCode: response.CodeSuccess, StatusMsg: resp.StatusCode.Msg()},
		NextTime:  resp.NextTime,
		VideoList: resp.VideoList,
	})
}
