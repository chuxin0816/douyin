package main

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"douyin/src/dal"
	"douyin/src/dal/model"
	favorite "douyin/src/kitex_gen/favorite"
	"douyin/src/pkg/kafka"
	"douyin/src/pkg/tracing"

	"github.com/cloudwego/kitex/pkg/klog"
	"go.opentelemetry.io/otel/codes"
)

// FavoriteServiceImpl implements the last service interface defined in the IDL.
type FavoriteServiceImpl struct{}

// FavoriteAction implements the FavoriteServiceImpl interface.
func (s *FavoriteServiceImpl) FavoriteAction(ctx context.Context, req *favorite.FavoriteActionRequest) (resp *favorite.FavoriteActionResponse, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "FavoriteAction")
	defer span.End()

	// 判断视频是否存在
	if err := dal.CheckVideoExist(ctx, req.VideoId); err != nil {
		span.RecordError(err)

		if errors.Is(err, dal.ErrVideoNotExist) {
			span.SetStatus(codes.Error, "视频不存在")
			klog.Error("视频不存在, videoID: ", req.VideoId)
			return nil, err
		}
		span.SetStatus(codes.Error, "判断视频是否存在失败")
		klog.Error("判断视频是否存在失败, err: ", err)
		return nil, err
	}

	// 获取作者ID
	authorID, err := dal.GetAuthorID(ctx, req.VideoId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "获取作者ID失败")
		klog.Error("获取作者ID失败, err: ", err)
		return nil, err
	}

	// 检查是否已经点赞
	exist, err := dal.CheckFavoriteExist(ctx, req.UserId, req.VideoId)
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

	// 通过kafka更新favorite表
	err = kafka.Favorite(ctx, &model.Favorite{
		ID:      req.ActionType,
		UserID:  req.UserId,
		VideoID: req.VideoId,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "通过kafka更新favorite表失败")
		klog.Error("通过kafka更新favorite表失败, err: ", err)
		return nil, err
	}

	keyVideoFavoriteCnt := dal.GetRedisKey(dal.KeyVideoFavoriteCountPF + strconv.FormatInt(req.VideoId, 10))
	keyUserFavoriteCnt := dal.GetRedisKey(dal.KeyUserFavoriteCountPF + strconv.FormatInt(req.UserId, 10))
	keyUserTotalFavorited := dal.GetRedisKey(dal.KeyUserTotalFavoritedPF + strconv.FormatInt(authorID, 10))

	// 检查相关字段是否存在缓存
	var wg sync.WaitGroup
	var errExist error
	wg.Add(3)
	go func() {
		defer wg.Done()
		if exist, err := dal.RDB.Exists(ctx, keyVideoFavoriteCnt).Result(); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "检查key是否存在失败")
			klog.Error("检查key是否存在失败, err: ", err)
			errExist = err
			return
		} else if exist == 0 {
			// 缓存不存在，查询数据库写入缓存
			cnt, err := dal.GetVideoFavoriteCount(ctx, req.VideoId)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "查询视频点赞数失败")
				klog.Error("查询数据库失败, err: ", err)
				errExist = err
				return
			}
			if err := dal.RDB.Set(ctx, keyVideoFavoriteCnt, cnt, 0).Err(); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "写入缓存失败")
				klog.Error("写入缓存失败, err: ", err)
				errExist = err
				return
			}
		}
	}()
	go func() {
		defer wg.Done()
		if exist, err := dal.RDB.Exists(ctx, keyUserFavoriteCnt).Result(); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "检查key是否存在失败")
			klog.Error("检查key是否存在失败, err: ", err)
			errExist = err
			return
		} else if exist == 0 {
			// 缓存不存在，查询数据库写入缓存
			cnt, err := dal.GetUserFavoriteCount(ctx, req.UserId)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "查询用户点赞数失败")
				klog.Error("查询数据库失败, err: ", err)
				errExist = err
				return
			}
			if err := dal.RDB.Set(ctx, keyUserFavoriteCnt, cnt, 0).Err(); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "写入缓存失败")
				klog.Error("写入缓存失败, err: ", err)
				errExist = err
				return
			}
		}
	}()
	go func() {
		defer wg.Done()
		if exist, err := dal.RDB.Exists(ctx, keyUserTotalFavorited).Result(); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "检查key是否存在失败")
			klog.Error("检查key是否存在失败, err: ", err)
			errExist = err
			return
		} else if exist == 0 {
			// 缓存不存在，查询数据库写入缓存
			cnt, err := dal.GetUserTotalFavorited(ctx, authorID)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "查询用户总点赞数失败")
				klog.Error("查询数据库失败, err: ", err)
				errExist = err
				return
			}
			if err := dal.RDB.Set(ctx, keyUserTotalFavorited, cnt, 0).Err(); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "写入缓存失败")
				klog.Error("写入缓存失败, err: ", err)
				errExist = err
				return
			}
		}
	}()
	wg.Wait()

	if errExist != nil {
		return nil, errExist
	}

	// 更新缓存
	pipe := dal.RDB.Pipeline()
	pipe.IncrBy(ctx, keyVideoFavoriteCnt, req.ActionType)
	pipe.IncrBy(ctx, keyUserFavoriteCnt, req.ActionType)
	pipe.IncrBy(ctx, keyUserTotalFavorited, req.ActionType)
	if _, err = pipe.Exec(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "更新缓存相关字段失败")
		klog.Error("更新缓存相关字段失败, err: ", err)
		return nil, err
	}

	// 写入待同步切片
	go func() {
		dal.Mu.Lock()
		dal.CacheUserID[req.UserId] = struct{}{}
		dal.CacheUserID[authorID] = struct{}{}
		dal.CacheVideoID[req.VideoId] = struct{}{}
		dal.Mu.Unlock()
	}()

	// 返回响应
	resp = &favorite.FavoriteActionResponse{}

	return
}

// FavoriteList implements the FavoriteServiceImpl interface.
func (s *FavoriteServiceImpl) FavoriteList(ctx context.Context, req *favorite.FavoriteListRequest) (resp *favorite.FavoriteListResponse, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "FavoriteList")
	defer span.End()

	// 获取喜欢的视频ID列表
	videoIDs, err := dal.GetFavoriteList(ctx, req.ToUserId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "获取喜欢的视频ID列表失败")
		klog.Error("获取喜欢的视频ID列表失败, err: ", err)
		return nil, err
	}

	// 获取视频列表
	videoList, err := dal.GetVideoList(ctx, req.UserId, videoIDs)
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
