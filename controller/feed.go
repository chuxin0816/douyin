package controller

import (
	"context"
	"douyin/pkg/jwt"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type FeedRequest struct {
	LatestTime int64  `query:"latest_time,string"` // 可选参数，限制返回视频的最新投稿时间戳，精确到秒，不填表示当前时间
	Token      string `query:"token"`              // 用户登录状态下设置
}

type FeedResponse struct {
	*Response
	VideoList []*VideoResponse `json:"video_list"` // 视频列表
	NextTime  *int64           `json:"next_time"`  // 本次返回的视频中，发布最早的时间，作为下次请求时的latest_time
}

// Feed 不限制登录状态，返回按投稿时间倒序的视频列表，视频数由服务端控制，单次最多30个
func Feed(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &FeedRequest{LatestTime: time.Now().Unix()}
	err := ctx.Bind(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
		hlog.Error("controller.Feed: 参数解析失败, err: ", err)
		return
	}

	// 验证token
	var userID int64
	if len(req.Token) > 0 {
		userID, err = jwt.ParseToken(req.Token)
		if err != nil {
			Error(ctx, CodeNoAuthority)
			hlog.Error("controller.Action: token无效, err: ", err)
			return
		}
	}

	// 业务逻辑处理
	resp, err := service.Feed(req.LatestTime, userID)
	if err != nil {
		Error(ctx, CodeServerBusy)
		hlog.Error("controller.Feed: 业务逻辑处理失败, err: ", err)
		return
	}

	// 返回结果
	Success(ctx, FeedResponse{
		Response:  &Response{StatusCode: CodeSuccess, StatusMsg: resp.StatusCode.Msg()},
		VideoList: resp.VideoList,
		NextTime:  resp.NextTime,
	})
}
