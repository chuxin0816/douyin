package main

import (
	"context"
	"douyin/config"
	"douyin/dal"
	feed "douyin/rpc/kitex_gen/feed"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

const count = 30

// FeedServiceImpl implements the last service interface defined in the IDL.
type FeedServiceImpl struct{}

// Feed implements the FeedServiceImpl interface.
func (s *FeedServiceImpl) Feed(ctx context.Context, req *feed.FeedRequest) (resp *feed.FeedResponse, err error) {
	ctx, span := otel.Tracer(config.Conf.OpenTelemetryConfig.FeedName).Start(ctx, "rpc.Feed")
	defer span.End()

	// 参数解析
	latestTime := time.Unix(req.LatestTime, 0)
	year := latestTime.Year()
	if year < 1 || year > 9999 {
		latestTime = time.Now()
	}

	// 查询视频列表
	videoList, nextTime, err := dal.GetFeedList(ctx, req.UserId, latestTime, count)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "查询视频列表失败")
		klog.Error("service.Feed: 查询视频列表失败")
		return nil, err
	}

	// 返回响应
	resp = &feed.FeedResponse{VideoList: videoList, NextTime: nextTime}

	return
}
