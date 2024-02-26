package main

import (
	"context"
	"douyin/dal"
	"douyin/dal/model"
	relation "douyin/rpc/kitex_gen/relation"
	"douyin/rpc/kitex_gen/user"

	"github.com/u2takey/go-utils/klog"
)

// RelationServiceImpl implements the last service interface defined in the IDL.
type RelationServiceImpl struct{}

// RelationAction implements the RelationServiceImpl interface.
func (s *RelationServiceImpl) RelationAction(ctx context.Context, req *relation.RelationActionRequest) (resp *relation.RelationActionResponse, err error) {
	// 操作数据库
	if err := dal.RelationAction(ctx, req.UserId, req.ToUserId, req.ActionType); err != nil {
		klog.Error("操作数据库失败, err: ", err)
		return nil, err
	}

	// 返回响应
	resp = &relation.RelationActionResponse{}

	return
}

// RelationFollowList implements the RelationServiceImpl interface.
func (s *RelationServiceImpl) RelationFollowList(ctx context.Context, req *relation.RelationFollowListRequest) (resp *relation.RelationFollowListResponse, err error) {
	// 操作数据库
	mUserList, err := dal.FollowList(ctx, req.ToUserId)
	if err != nil {
		klog.Error("操作数据库失败, err: ", err)
		return nil, err
	}

	//将model.User转换为user.User
	userList := make([]*user.User, len(mUserList))
	for i, u := range mUserList {
		userList[i] = dal.ToUserResponse(ctx, req.UserId, u)
	}

	// 返回响应
	resp = &relation.RelationFollowListResponse{UserList: userList}

	return
}

// RelationFollowerList implements the RelationServiceImpl interface.
func (s *RelationServiceImpl) RelationFollowerList(ctx context.Context, req *relation.RelationFollowerListRequest) (resp *relation.RelationFollowerListResponse, err error) {
	// 操作数据库
	mUserList, err := dal.FollowerList(ctx, req.ToUserId)
	if err != nil {
		klog.Error("操作数据库失败, err: ", err)
		return nil, err
	}

	// 将model.User转换为user.User
	userList := make([]*user.User, len(mUserList))
	for i, u := range mUserList {
		userList[i] = dal.ToUserResponse(ctx, req.UserId, u)
	}

	// 返回响应
	resp = &relation.RelationFollowerListResponse{UserList: userList}

	return
}

// RelationFriendList implements the RelationServiceImpl interface.
func (s *RelationServiceImpl) RelationFriendList(ctx context.Context, req *relation.RelationFriendListRequest) (resp *relation.RelationFriendListResponse, err error) {
	// 获取关注列表
	mFollowList, err := dal.FollowList(ctx, req.ToUserId)
	if err != nil {
		klog.Error("获取关注列表失败, err: ", err)
		return nil, err
	}

	// 获取粉丝列表
	mFollowerList, err := dal.FollowerList(ctx, req.ToUserId)
	if err != nil {
		klog.Error("获取粉丝列表失败, err: ", err)
		return nil, err
	}

	// 获取好友列表
	size := min(len(mFollowList), len(mFollowerList))
	dFriendList := make([]*model.User, 0, size)
	mp := make(map[int64]struct{}, len(mFollowList))
	for _, user := range mFollowList {
		mp[user.ID] = struct{}{}
	}

	for _, user := range mFollowerList {
		if _, ok := mp[user.ID]; ok {
			dFriendList = append(dFriendList, user)
		}
	}

	// 将model.User转换为user.User
	friendList := make([]*user.User, len(dFriendList))
	for i, u := range dFriendList {
		friendList[i] = dal.ToUserResponse(ctx, req.UserId, u)
	}

	// 返回响应
	resp = &relation.RelationFriendListResponse{UserList: friendList}

	return
}
