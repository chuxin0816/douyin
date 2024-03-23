package main

import (
	"context"
	"douyin/dal"
	"douyin/dal/model"
	"douyin/pkg/kafka"
	"douyin/pkg/tracing"
	favorite "douyin/rpc/kitex_gen/favorite"
	"errors"
	"strconv"

	"github.com/cloudwego/kitex/pkg/klog"
	"go.opentelemetry.io/otel/codes"
)

// FavoriteServiceImpl implements the last service interface defined in the IDL.
type FavoriteServiceImpl struct{}

// FavoriteAction implements the FavoriteServiceImpl interface.
func (s *FavoriteServiceImpl) FavoriteAction(ctx context.Context, req *favorite.FavoriteActionRequest) (resp *favorite.FavoriteActionResponse, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "rpc.FavoriteAction")
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
	err = kafka.Favorite(&model.Favorite{
		UserID:  req.UserId,
		VideoID: req.VideoId,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "通过kafka更新favorite表失败")
		klog.Error("通过kafka更新favorite表失败, err: ", err)
		return nil, err
	}

	// 更新缓存相关字段
	go func() {
		keyVideoFavoriteCnt := dal.GetRedisKey(dal.KeyVideoFavoriteCountPF + strconv.FormatInt(req.VideoId, 10))
		keyUserFavoriteCnt := dal.GetRedisKey(dal.KeyUserFavoriteCountPF + strconv.FormatInt(req.UserId, 10))
		keyUserTotalFavorited := dal.GetRedisKey(dal.KeyUserTotalFavoritedPF + strconv.FormatInt(authorID, 10))
		// 检查key是否存在
		if exist, err := dal.RDB.Exists(ctx, keyVideoFavoriteCnt, keyUserFavoriteCnt, keyUserTotalFavorited).Result(); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "检查key是否存在失败")
			klog.Error("检查key是否存在失败, err: ", err)
			return
		} else if exist != 3 {
			// 缓存不存在，查询数据库写入缓存
			cnt, err := dal.GetVideoFavoriteCount(ctx, req.VideoId)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "查询视频点赞数失败")
				klog.Error("查询数据库失败, err: ", err)
				return
			}
			if err := dal.RDB.Set(ctx, keyVideoFavoriteCnt, cnt, 0).Err(); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "写入缓存失败")
				klog.Error("写入缓存失败, err: ", err)
				return
			}
			cnt, err = dal.GetUserFavoriteCount(ctx, req.UserId)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "查询用户点赞数失败")
				klog.Error("查询数据库失败, err: ", err)
				return
			}
			if err := dal.RDB.Set(ctx, keyUserFavoriteCnt, cnt, 0).Err(); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "写入缓存失败")
				klog.Error("写入缓存失败, err: ", err)
				return
			}
			cnt, err = dal.GetUserTotalFavorited(ctx, authorID)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "查询用户总点赞数失败")
				klog.Error("查询数据库失败, err: ", err)
				return
			}
			if err := dal.RDB.Set(ctx, keyUserTotalFavorited, cnt, 0).Err(); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "写入缓存失败")
				klog.Error("写入缓存失败, err: ", err)
				return
			}
		}

		pipe := dal.RDB.Pipeline()
		// 更新video的favorite_count字段
		pipe.IncrBy(ctx, keyVideoFavoriteCnt, req.ActionType)

		// 更新user当前用户的favorite_count字段
		pipe.IncrBy(ctx, keyUserFavoriteCnt, req.ActionType)

		// 更新user作者的total_favorited字段
		pipe.IncrBy(ctx, keyUserTotalFavorited, req.ActionType)

		if _, err := pipe.Exec(ctx); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "更新缓存相关字段失败")
			klog.Error("更新缓存相关字段失败, err: ", err)
			return
		}

		// 写入待同步切片
		dal.CacheUserID.Store(req.UserId, struct{}{})
		dal.CacheUserID.Store(authorID, struct{}{})
		dal.CacheVideoID.Store(req.VideoId, struct{}{})
	}()

	// 返回响应
	resp = &favorite.FavoriteActionResponse{}

	return
}

// FavoriteList implements the FavoriteServiceImpl interface.
func (s *FavoriteServiceImpl) FavoriteList(ctx context.Context, req *favorite.FavoriteListRequest) (resp *favorite.FavoriteListResponse, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "rpc.FavoriteList")
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
