package main

import (
	"context"
	"strconv"
	"sync/atomic"
	"time"

	"douyin/src/client"
	"douyin/src/common/kafka"
	"douyin/src/dal"
	"douyin/src/kitex_gen/favorite"
	"douyin/src/kitex_gen/video"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

// FavoriteServiceImpl implements the last service interface defined in the IDL.
type FavoriteServiceImpl struct{}

// FavoriteAction implements the FavoriteServiceImpl interface.
func (s *FavoriteServiceImpl) FavoriteAction(ctx context.Context, req *favorite.FavoriteActionRequest) (resp *favorite.FavoriteActionResponse, err error) {
	ctx, span := otel.Tracer("favorite").Start(ctx, "FavoriteAction")
	defer span.End()

	// 判断视频是否存在
	exist, err := client.VideoClient.VideoExist(ctx, req.VideoId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "查询视频失败")
		klog.Error("查询视频失败, err: ", err)
		return nil, err
	}
	if !exist {
		span.SetStatus(codes.Error, "视频不存在")
		klog.Error("视频不存在, err: ", err)
		return nil, dal.ErrVideoNotExist
	}

	// 获取作者ID
	authorID, err := client.VideoClient.AuthorId(ctx, req.VideoId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "获取作者ID失败")
		klog.Error("获取作者ID失败, err: ", err)
		return nil, err
	}

	// 检查是否已经点赞
	exist, err = dal.CheckFavoriteExist(ctx, req.UserId, req.VideoId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "检查是否已经点赞失败")
		klog.Error("检查是否已经点赞失败, err: ", err)
		return nil, err
	}

	// 已经点赞
	if exist && req.ActionType == 1 {
		return nil, dal.ErrAlreadyFavorite
	}
	// 未点赞
	if !exist && req.ActionType == -1 {
		return nil, dal.ErrNotFavorite
	}

	// 添加到map
	e := kafka.FavoriteEvent{
		UserID:     req.UserId,
		VideoID:    req.VideoId,
		ActionType: req.ActionType,
	}
	kafka.FavoriteMap.Set(e.String(), e)

	keyUserFavorite := dal.GetRedisKey(dal.KeyUserFavoritePF, strconv.FormatInt(req.UserId, 10))
	keyVideoFavoriteCnt := dal.GetRedisKey(dal.KeyVideoFavoriteCountPF, strconv.FormatInt(req.VideoId, 10))
	keyUserFavoriteCnt := dal.GetRedisKey(dal.KeyUserFavoriteCountPF, strconv.FormatInt(req.UserId, 10))
	keyUserTotalFavorited := dal.GetRedisKey(dal.KeyUserTotalFavoritedPF, strconv.FormatInt(authorID, 10))

	// 更新缓存
	pipe := dal.RDB.Pipeline()
	if req.ActionType == 1 {
		pipe.SAdd(ctx, keyUserFavorite, req.VideoId)
		pipe.Expire(ctx, keyUserFavorite, dal.ExpireTime+dal.GetRandomTime())
	} else {
		pipe.SRem(ctx, keyUserFavorite, req.VideoId)
	}
	dal.IncrByScript.Run(ctx, pipe, []string{keyVideoFavoriteCnt}, req.ActionType)
	dal.IncrByScript.Run(ctx, pipe, []string{keyUserFavoriteCnt}, req.ActionType)
	dal.IncrByScript.Run(ctx, pipe, []string{keyUserTotalFavorited}, req.ActionType)
	if _, err = pipe.Exec(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "更新缓存相关字段失败")
		klog.Error("更新缓存相关字段失败, err: ", err)
		return nil, err
	}

	// 返回响应
	resp = &favorite.FavoriteActionResponse{}

	return
}

// FavoriteList implements the FavoriteServiceImpl interface.
func (s *FavoriteServiceImpl) FavoriteList(ctx context.Context, req *favorite.FavoriteListRequest) (resp *favorite.FavoriteListResponse, err error) {
	ctx, span := otel.Tracer("favorite").Start(ctx, "FavoriteList")
	defer span.End()

	// 获取喜欢的视频ID列表
	videoIDs, err := dal.GetFavoriteList(ctx, req.AuthorId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "获取喜欢的视频ID列表失败")
		klog.Error("获取喜欢的视频ID列表失败, err: ", err)
		return nil, err
	}

	// 获取视频列表
	videoList, err := client.VideoClient.VideoInfoList(ctx, &video.VideoInfoListRequest{
		UserId:      req.UserId,
		VideoIdList: videoIDs,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "获取视频列表失败")
		klog.Error("获取视频列表失败, err: ", err)
		return nil, err
	}
	// 返回响应
	resp = &favorite.FavoriteListResponse{VideoList: videoList}

	return
}

// FavoriteCnt implements the FavoriteServiceImpl interface.
func (s *FavoriteServiceImpl) FavoriteCnt(ctx context.Context, userId int64) (resp int64, err error) {
	ctx, span := otel.Tracer("favorite").Start(ctx, "FavoriteCnt")
	defer span.End()

	resp, err = dal.GetUserFavoriteCount(ctx, userId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "获取用户喜欢的视频数失败")
		klog.Error("获取用户喜欢的视频数失败, err: ", err)
		return
	}

	return
}

// TotalFavoritedCnt implements the FavoriteServiceImpl interface.
func (s *FavoriteServiceImpl) TotalFavoritedCnt(ctx context.Context, userId int64) (resp int64, err error) {
	ctx, span := otel.Tracer("favorite").Start(ctx, "TotalFavoritedCnt")
	defer span.End()

	key := dal.GetRedisKey(dal.KeyUserTotalFavoritedPF, strconv.FormatInt(userId, 10))
	// 使用singleflight解决缓存击穿并减少redis压力
	_, err, _ = dal.G.Do(key, func() (interface{}, error) {
		go func() {
			time.Sleep(dal.DelayTime)
			dal.G.Forget(key)
		}()
		// 先查询redis缓存
		resp, err = dal.RDB.Get(ctx, key).Int64()
		if err == redis.Nil {
			// 缓存未命中，查询mysql
			// 查询用户发布列表
			videoIDs, err := client.VideoClient.PublishIDList(ctx, userId)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "查询用户发布列表失败")
				klog.Error("查询用户发布列表失败, err: ", err)
				return nil, err
			}

			// 查询用户发布视频的点赞数
			for _, videoID := range videoIDs {
				cnt, err := dal.GetVideoFavoriteCount(ctx, videoID)
				if err != nil {
					span.RecordError(err)
					span.SetStatus(codes.Error, "查询用户发布视频的点赞数失败")
					klog.Error("查询用户发布视频的点赞数失败, err: ", err)
					return nil, err
				}
				atomic.AddInt64(&resp, cnt)
			}

			// 写入redis缓存
			dal.RDB.Set(ctx, key, resp, dal.ExpireTime+dal.GetRandomTime())

		} else if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "查询用户总点赞数失败")
			klog.Error("查询用户总点赞数失败, err: ", err)
			return nil, err
		}

		return nil, nil
	})

	return
}

// FavoriteExist implements the FavoriteServiceImpl interface.
func (s *FavoriteServiceImpl) FavoriteExist(ctx context.Context, userId int64, videoId int64) (resp bool, err error) {
	ctx, span := otel.Tracer("favorite").Start(ctx, "FavoriteExist")
	defer span.End()

	resp, err = dal.CheckFavoriteExist(ctx, userId, videoId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "查询是否已经点赞失败")
		klog.Error("查询是否已经点赞失败, err: ", err)
		return
	}

	return
}
