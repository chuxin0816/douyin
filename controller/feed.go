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
)

// Feed 不限制登录状态，返回按投稿时间倒序的视频列表，视频数由服务端控制，单次最多30个
func Feed(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &models.FeedRequest{LatestTime: strconv.FormatInt(time.Now().Unix(), 10)}
	err := ctx.Bind(req)
	if err != nil {
		response.Error(ctx, response.CodeInvalidParam)
		hlog.Error("controller.Feed: 参数解析失败, err: ", err)
		return
	}

	// 业务逻辑处理
	resp, err := service.Feed(req)
	if err != nil {
		response.Error(ctx, response.CodeServerBusy)
		hlog.Error("controller.Feed: 业务逻辑处理失败, err: ", err)
		return
	}

	// 返回结果
	response.Success(ctx, response.FeedResponse{
		Response:  &response.Response{StatusCode: response.CodeSuccess, StatusMsg: resp.StatusCode.Msg()},
		VideoList: resp.VideoList,
		NextTime:  resp.NextTime,
	})
}
