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

func FollowList(userID, toUserID int64) (*response.FollowListResponse, error) {
	// 操作数据库
	dUserList, err := mysql.FollowList(userID, toUserID)
	if err != nil {
		hlog.Error("service.FollowList: 操作数据库失败, err: ", err)
		return nil, err
	}

	// 将models.User转换为response.UserResponse
	userList := make([]*response.UserResponse, 0, len(dUserList))
	for _, user := range dUserList {
		userList = append(userList, response.ToUserResponse(user))
	}

	// 返回响应
	return &response.FollowListResponse{
		Response: &response.Response{StatusCode: response.CodeSuccess, StatusMsg: response.CodeSuccess.Msg()},
		UserList: userList,
	}, nil
}

func FollowerList(userID, toUserID int64) (*response.FollowerListResponse, error) {
	// 操作数据库
	dUserList, err := mysql.FollowerList(userID, toUserID)
	if err != nil {
		hlog.Error("service.FollowerList: 操作数据库失败, err: ", err)
		return nil, err
	}

	// 将models.User转换为response.UserResponse
	userList := make([]*response.UserResponse, 0, len(dUserList))
	for _, user := range dUserList {
		userList = append(userList, response.ToUserResponse(user))
	}

	// 返回响应
	return &response.FollowerListResponse{
		Response: &response.Response{StatusCode: response.CodeSuccess, StatusMsg: response.CodeSuccess.Msg()},
		UserList: userList,
	}, nil
}
