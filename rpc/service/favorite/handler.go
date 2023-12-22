package main

import (
	"context"
	"douyin/dal"
	favorite "douyin/rpc/kitex_gen/favorite"

	"github.com/cloudwego/kitex/pkg/klog"
)

// FavoriteServiceImpl implements the last service interface defined in the IDL.
type FavoriteServiceImpl struct{}

// FavoriteAction implements the FavoriteServiceImpl interface.
func (s *FavoriteServiceImpl) FavoriteAction(ctx context.Context, req *favorite.FavoriteActionRequest) (resp *favorite.FavoriteActionResponse, err error) {
	// 解析视频点赞类型
	if req.ActionType == 2 {
		req.ActionType = -1
	}

	// 操作数据库
	if err := dal.FavoriteAction(req.UserId, req.VideoId, req.ActionType); err != nil {
		klog.Error("操作数据库失败, err: ", err)
		return nil, err
	}

	// 返回响应
	resp = &favorite.FavoriteActionResponse{}

	return
}

// FavoriteList implements the FavoriteServiceImpl interface.
func (s *FavoriteServiceImpl) FavoriteList(ctx context.Context, req *favorite.FavoriteListRequest) (resp *favorite.FavoriteListResponse, err error) {
	// 获取喜欢的视频ID列表
	videoIDs, err := dal.GetFavoriteList(req.UserId)
	if err != nil {
		klog.Error("获取喜欢的视频ID列表失败, err: ", err)
		return nil, err
	}

	// 获取视频列表
	videoList, err := dal.GetVideoList(req.UserId, videoIDs)
	if err != nil {
		klog.Error("获取视频列表失败, err: ", err)
		return nil, err
	}

	// 返回响应
	resp = &favorite.FavoriteListResponse{VideoList: videoList}

	return
}
