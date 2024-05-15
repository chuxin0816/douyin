package main

import (
	"context"
	"sync"
	"time"

	"douyin/src/dal"
	"douyin/src/dal/model"
	feed "douyin/src/kitex_gen/feed"
	"douyin/src/pkg/tracing"

	"github.com/cloudwego/kitex/pkg/klog"
	"go.opentelemetry.io/otel/codes"
)

const count = 30

// FeedServiceImpl implements the last service interface defined in the IDL.
type FeedServiceImpl struct{}

// Feed implements the FeedServiceImpl interface.
func (s *FeedServiceImpl) Feed(ctx context.Context, req *feed.FeedRequest) (resp *feed.FeedResponse, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "Feed")
	defer span.End()

	// 参数解析
	latestTime := time.Unix(req.LatestTime, 0)
	year := latestTime.Year()
	if year < 1 || year > 9999 {
		latestTime = time.Now()
	}

	// 查询视频列表
	mVideoList, err := dal.GetFeedList(ctx, req.UserId, latestTime, count)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "查询视频列表失败")
		klog.Error("service.Feed: 查询视频列表失败, err: ", err)
		return nil, err
	}

	// 将model.Video转换为feed.Video
	videoList := make([]*feed.Video, len(mVideoList))
	var wg sync.WaitGroup
	wg.Add(len(mVideoList))
	for i, mVideo := range mVideoList {
		go func(i int, mVideo *model.Video) {
			defer wg.Done()
			videoList[i] = dal.ToVideoResponse(ctx, req.UserId, mVideo)
		}(i, mVideo)
	}
	wg.Wait()

	// 计算下次请求的时间
	var nextTime *int64
	if len(mVideoList) > 0 {
		nextTime = new(int64)
		*nextTime = mVideoList[len(mVideoList)-1].UploadTime.Unix()
	}

	// 返回响应
	resp = &feed.FeedResponse{VideoList: videoList, NextTime: nextTime}

	return
}
