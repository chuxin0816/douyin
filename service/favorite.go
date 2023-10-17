package service

import (
	"douyin/dao/mysql"
	"douyin/response"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func FavoriteAction(userID int64, videoID int64, actionType int) (*response.FavoriteActionResponse, error) {
	// 解析视频点赞类型
	if actionType == 2 {
		actionType = -1
	}

	// 操作数据库
	err := mysql.FavoriteAction(userID, videoID, actionType)
	if err != nil {
		hlog.Error("service.FavoriteAction: 操作数据库失败, err: ", err)
		return nil, err
	}

	// 返回响应
	return &response.FavoriteActionResponse{
		Response: &response.Response{StatusCode: response.CodeSuccess, StatusMsg: response.CodeSuccess.Msg()},
	}, nil
}
