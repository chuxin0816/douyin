package service

import (
	"douyin/dao"
	"douyin/response"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func FavoriteAction(userID int64, videoID int64, actionType int64) (*response.FavoriteActionResponse, error) {
	// 解析视频点赞类型
	if actionType == 2 {
		actionType = -1
	}

	// 操作数据库
	if err := dao.FavoriteAction(userID, videoID, actionType); err != nil {
		hlog.Error("service.FavoriteAction: 操作数据库失败, err: ", err)
		return nil, err
	}

	// 返回响应
	return &response.FavoriteActionResponse{
		Response: &response.Response{StatusCode: response.CodeSuccess, StatusMsg: response.CodeSuccess.Msg()},
	}, nil
}

func FavoriteList(userID, authorID int64) (*response.FavoriteListResponse, error) {
	// 获取喜欢的视频ID列表
	videoIDs, err := dao.GetFavoriteList(authorID)
	if err != nil {
		hlog.Error("service.FavoriteList: 获取喜欢的视频ID列表失败, err: ", err)
		return nil, err
	}

	// 获取视频列表
	videoList, err := dao.GetVideoList(userID, videoIDs)
	if err != nil {
		hlog.Error("service.FavoriteList: 获取视频列表失败, err: ", err)
		return nil, err
	}

	// 返回响应
	return &response.FavoriteListResponse{
		Response:  &response.Response{StatusCode: response.CodeSuccess, StatusMsg: response.CodeSuccess.Msg()},
		VideoList: videoList,
	}, nil
}
