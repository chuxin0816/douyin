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
	for i, mVideo := range mVideoList {
		videoList[i] = toVideoResponse(ctx, req.UserId, mVideo)
	}

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

func toVideoResponse(ctx context.Context, userID *int64, mVideo *model.Video) *feed.Video {
	video := &feed.Video{
		Id:         mVideo.ID,
		PlayUrl:    mVideo.PlayURL,
		CoverUrl:   mVideo.CoverURL,
		IsFavorite: false,
		Title:      mVideo.Title,
	}

	var wg sync.WaitGroup
	var wgErr error
	wg.Add(3)
	go func() {
		defer wg.Done()
		author, err := dal.GetUserByID(ctx, mVideo.AuthorID)
		if err != nil {
			wgErr = err
			return
		}
		video.Author = ToUserResponse(ctx, userID, author)
	}()
	go func() {
		defer wg.Done()
		cnt, err := dal.GetVideoCommentCount(ctx, mVideo.ID)
		if err != nil {
			wgErr = err
			return
		}
		video.CommentCount = cnt
	}()
	go func() {
		defer wg.Done()
		cnt, err := GetVideoFavoriteCount(ctx, mVideo.ID)
		if err != nil {
			wgErr = err
			return
		}
		video.FavoriteCount = cnt
	}()
	wg.Wait()
	if wgErr != nil {
		return video
	}

	// 未登录直接返回
	if userID == nil || *userID == 0 {
		return video
	}

	// 查询缓存判断是否点赞
	exist, err := CheckFavoriteExist(ctx, *userID, mVideo.ID)
	if err != nil {
		return video
	}
	video.IsFavorite = exist

	return video
}
