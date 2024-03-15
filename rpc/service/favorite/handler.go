package main

import (
	"context"
	"douyin/dal"
	"douyin/dal/model"
	"douyin/pkg/kafka"
	favorite "douyin/rpc/kitex_gen/favorite"
	"errors"
	"strconv"

	"github.com/cloudwego/kitex/pkg/klog"
)

// FavoriteServiceImpl implements the last service interface defined in the IDL.
type FavoriteServiceImpl struct{}

// FavoriteAction implements the FavoriteServiceImpl interface.
func (s *FavoriteServiceImpl) FavoriteAction(ctx context.Context, req *favorite.FavoriteActionRequest) (resp *favorite.FavoriteActionResponse, err error) {
	// 判断视频是否存在
	if err := dal.CheckVideoExist(ctx, req.VideoId); err != nil {
		if errors.Is(err, dal.ErrVideoNotExist) {
			klog.Error("视频不存在, videoID: ", req.VideoId)
			return nil, err
		}
		klog.Error("判断视频是否存在失败, err: ", err)
		return nil, err
	}

	// 获取作者ID
	authorID, err := dal.GetAuthorID(ctx, req.VideoId)
	if err != nil {
		klog.Error("获取作者ID失败, err: ", err)
		return nil, err
	}

	// 检查是否已经点赞
	exist, err := dal.CheckFavoriteExist(ctx, req.UserId, req.VideoId)
	if err != nil {
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
		klog.Error("通过kafka更新favorite表失败, err: ", err)
		return nil, err
	}

	// 更新缓存相关字段
	go func() {
		// 检查key是否存在
		keyVideoFavoriteCnt := dal.GetRedisKey(dal.KeyVideoFavoriteCountPF + strconv.FormatInt(req.VideoId, 10))
		keyUserFavoriteCnt := dal.GetRedisKey(dal.KeyUserFavoriteCountPF + strconv.FormatInt(req.UserId, 10))
		keyUserTotalFavorited := dal.GetRedisKey(dal.KeyUserTotalFavoritedPF + strconv.FormatInt(authorID, 10))
		

		pipe := dal.RDB.Pipeline()
		// 更新video的favorite_count字段
		pipe.IncrBy(ctx, keyVideoFavoriteCnt, req.ActionType)

		// 更新user当前用户的favorite_count字段
		pipe.IncrBy(ctx, keyUserFavoriteCnt, req.ActionType)

		// 更新user作者的total_favorited字段
		pipe.IncrBy(ctx, keyUserTotalFavorited, req.ActionType)

		if _, err := pipe.Exec(ctx); err != nil {
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
	// 获取喜欢的视频ID列表
	videoIDs, err := dal.GetFavoriteList(ctx, req.ToUserId)
	if err != nil {
		klog.Error("获取喜欢的视频ID列表失败, err: ", err)
		return nil, err
	}

	// 获取视频列表
	videoList, err := dal.GetVideoList(ctx, req.UserId, videoIDs)
	if err != nil {
		klog.Error("获取视频列表失败, err: ", err)
		return nil, err
	}

	// 返回响应
	resp = &favorite.FavoriteListResponse{VideoList: videoList}

	return
}
