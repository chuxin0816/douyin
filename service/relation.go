package service

import (
	"douyin/dao/mysql"
	"douyin/models"
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

func FollowList(userID, toUserID int64) (*response.RelationListResponse, error) {
	// 操作数据库
	dUserList, err := mysql.FollowList(toUserID)
	if err != nil {
		hlog.Error("service.FollowList: 操作数据库失败, err: ", err)
		return nil, err
	}

	// 将models.User转换为response.UserResponse
	userList := make([]*response.UserResponse, 0, len(dUserList))
	for _, user := range dUserList {
		userList = append(userList, mysql.ToUserResponse(userID, user))
	}

	// 返回响应
	return &response.RelationListResponse{
		Response: &response.Response{StatusCode: response.CodeSuccess, StatusMsg: response.CodeSuccess.Msg()},
		UserList: userList,
	}, nil
}

func FollowerList(userID, toUserID int64) (*response.RelationListResponse, error) {
	// 操作数据库
	dUserList, err := mysql.FollowerList(toUserID)
	if err != nil {
		hlog.Error("service.FollowerList: 操作数据库失败, err: ", err)
		return nil, err
	}

	// 将models.User转换为response.UserResponse
	userList := make([]*response.UserResponse, 0, len(dUserList))
	for _, user := range dUserList {
		userList = append(userList, mysql.ToUserResponse(userID, user))
	}

	// 返回响应
	return &response.RelationListResponse{
		Response: &response.Response{StatusCode: response.CodeSuccess, StatusMsg: response.CodeSuccess.Msg()},
		UserList: userList,
	}, nil
}

func FriendList(userID, toUserID int64) (*response.RelationListResponse, error) {
	// 获取关注列表
	dFollowList, err := mysql.FollowList(toUserID)
	if err != nil {
		hlog.Error("service.FriendList: 获取关注列表失败, err: ", err)
		return nil, err
	}

	// 获取粉丝列表
	dFollowerList, err := mysql.FollowerList(toUserID)
	if err != nil {
		hlog.Error("service.FriendList: 获取粉丝列表失败, err: ", err)
		return nil, err
	}

	// 获取好友列表
	size := len(dFollowList)
	if size > len(dFollowerList) {
		size = len(dFollowerList)
	}
	dFriendList := make([]*models.User, 0, size)
	mp := make(map[*models.User]struct{}, len(dFollowList))
	for _, user := range dFollowList {
		mp[user] = struct{}{}
	}
	for _, user := range dFollowerList {
		if _, ok := mp[user]; ok {
			dFriendList = append(dFriendList, user)
		}
	}

	// 将models.User转换为response.UserResponse
	friendList := make([]*response.UserResponse, 0, len(dFriendList))
	for _, user := range dFriendList {
		friendList = append(friendList, mysql.ToUserResponse(userID, user))
	}

	// 返回响应
	return &response.RelationListResponse{
		Response: &response.Response{StatusCode: response.CodeSuccess, StatusMsg: response.CodeSuccess.Msg()},
		UserList: friendList,
	}, nil
}
