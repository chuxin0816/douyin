package main

import (
	"context"
	"douyin/dal"
	feed "douyin/rpc/kitex_gen/feed"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

const count = 30

// FeedServiceImpl implements the last service interface defined in the IDL.
type FeedServiceImpl struct{}

// Feed implements the FeedServiceImpl interface.
func (s *FeedServiceImpl) Feed(ctx context.Context, req *feed.FeedRequest) (resp *feed.FeedResponse, err error) {
	// 参数解析
	latestTime := time.Unix(req.LatestTime, 0)
	year := latestTime.Year()
	if year < 1 || year > 9999 {
		latestTime = time.Now()
	}

	// 查询视频列表
	videoList, nextTime, err := dal.GetFeedList(ctx, req.UserId, latestTime, count)
	if err != nil {
		hlog.Error("service.Feed: 查询视频列表失败")
		return nil, err
	}

	// 返回响应
	resp = &feed.FeedResponse{VideoList: videoList, NextTime: nextTime}

	return
}
