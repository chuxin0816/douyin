package controller

import (
	"context"
	"time"

	"douyin/src/pkg/jwt"
	"douyin/src/pkg/tracing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"go.opentelemetry.io/otel/codes"

	"douyin/src/config"
	"douyin/src/kitex_gen/feed"
	"douyin/src/kitex_gen/feed/feedservice"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	tracing2 "github.com/kitex-contrib/obs-opentelemetry/tracing"
	consul "github.com/kitex-contrib/registry-consul"
)

type FeedRequest struct {
	LatestTime int64  `query:"latest_time,string"` // 可选参数，限制返回视频的最新投稿时间戳，精确到秒，不填表示当前时间
	Token      string `query:"token"`              // 用户登录状态下设置
}

var feedClient feedservice.Client

func init() {
	// 服务发现
	r, err := consul.NewConsulResolver(config.Conf.ConsulConfig.ConsulAddr)
	if err != nil {
		panic(err)
	}

	feedClient, err = feedservice.NewClient(
		config.Conf.OpenTelemetryConfig.FeedName,
		client.WithResolver(r),
		client.WithSuite(tracing2.NewClientSuite()),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: config.Conf.OpenTelemetryConfig.ApiName}),
		client.WithMuxConnection(2),
	)
	if err != nil {
		panic(err)
	}
}

// Feed 不限制登录状态，返回按投稿时间倒序的视频列表，视频数由服务端控制，单次最多30个
func Feed(c context.Context, ctx *app.RequestContext) {
	c, span := tracing.Tracer.Start(c, "Feed")
	defer span.End()

	// 获取参数
	req := &FeedRequest{LatestTime: time.Now().Unix()}
	err := ctx.Bind(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
		span.RecordError(err)
		span.SetStatus(codes.Error, "参数解析失败")
		hlog.Error("参数解析失败, err: ", err)
		return
	}

	// 验证token
	userID := jwt.ParseToken(req.Token)

	// 业务逻辑处理
	resp, err := feedClient.Feed(c, &feed.FeedRequest{
		LatestTime: req.LatestTime,
		UserId:     userID,
	})
	if err != nil {
		Error(ctx, CodeServerBusy)
		span.RecordError(err)
		span.SetStatus(codes.Error, "业务逻辑处理失败")
		hlog.Error("业务逻辑处理失败, err: ", err)
		return
	}

	// 返回结果
	Success(ctx, resp)
}
