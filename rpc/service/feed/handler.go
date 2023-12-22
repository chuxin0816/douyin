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
	// 解析请求
	if req.LatestTime == nil {
		currentTime := time.Now().Unix()
		req.LatestTime = &currentTime
	}

	// 查询视频列表
	videoList, nextTime, err := dal.GetFeedList(req.UserId, time.Unix(*req.LatestTime, 0), count)
	if err != nil {
		hlog.Error("service.Feed: 查询视频列表失败")
		return nil, err
	}

	// 返回响应
	resp = &feed.FeedResponse{VideoList: videoList, NextTime: nextTime}

	return
}
