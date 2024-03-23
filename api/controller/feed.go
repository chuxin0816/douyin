package controller

import (
	"context"
	"douyin/pkg/jwt"
	"douyin/pkg/tracing"
	"douyin/rpc/client"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/kitex/pkg/klog"
	"go.opentelemetry.io/otel/codes"
)

type FeedRequest struct {
	LatestTime int64  `query:"latest_time,string"` // 可选参数，限制返回视频的最新投稿时间戳，精确到秒，不填表示当前时间
	Token      string `query:"token"`              // 用户登录状态下设置
}

// Feed 不限制登录状态，返回按投稿时间倒序的视频列表，视频数由服务端控制，单次最多30个
func Feed(c context.Context, ctx *app.RequestContext) {
	_, span := tracing.Tracer.Start(c, "controller.Feed")
	defer span.End()

	// 获取参数
	req := &FeedRequest{LatestTime: time.Now().Unix()}
	err := ctx.Bind(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
		span.RecordError(err)
		span.SetStatus(codes.Error, "参数解析失败")
		klog.Error("参数解析失败, err: ", err)
		return
	}

	// 验证token
	userID := jwt.ParseToken(req.Token)

	// 业务逻辑处理
	resp, err := client.Feed(req.LatestTime, userID)
	if err != nil {
		Error(ctx, CodeServerBusy)
		span.RecordError(err)
		span.SetStatus(codes.Error, "业务逻辑处理失败")
		klog.Error("业务逻辑处理失败, err: ", err)
		return
	}

	// 返回结果
	Success(ctx, resp)
}
