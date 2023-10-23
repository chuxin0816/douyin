package service

import (
	"douyin/dao/mysql"
	"douyin/response"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func RelationAction(userID, toUserID int64, actionType int) (*response.RelationActionResponse, error) {
	// 解析关注类型
	if actionType == 2 {
		actionType = -1
	}

	// 操作数据库
	err := mysql.RelationAction(userID, toUserID, actionType)
	if err != nil {
		hlog.Error("service.RelationAction: 操作数据库失败, err: ", err)
		return nil, err
	}
	
	// 返回响应
	return &response.RelationActionResponse{
		Response: &response.Response{StatusCode: response.CodeSuccess, StatusMsg: response.CodeSuccess.Msg()},
	}, nil
}
