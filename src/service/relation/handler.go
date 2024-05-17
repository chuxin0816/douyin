package main

import (
	"context"
	"sync"

	"douyin/src/dal"
	"douyin/src/dal/model"
	relation "douyin/src/kitex_gen/relation"
	"douyin/src/kitex_gen/user"
	"douyin/src/pkg/kafka"
	"douyin/src/pkg/tracing"

	"github.com/cloudwego/kitex/pkg/klog"
	"go.opentelemetry.io/otel/codes"
)

// RelationServiceImpl implements the last service interface defined in the IDL.
type RelationServiceImpl struct{}

// RelationAction implements the RelationServiceImpl interface.
func (s *RelationServiceImpl) RelationAction(ctx context.Context, req *relation.RelationActionRequest) (resp *relation.RelationActionResponse, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "RelationAction")
	defer span.End()

	// 检查是否关注
	exist, err := dal.CheckRelationExist(ctx, req.UserId, req.ToUserId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "检查是否关注失败")
		klog.Error("检查是否关注失败, err: ", err)
		return nil, err
	}

	if exist && req.ActionType == 1 {
		return nil, dal.ErrAlreadyFollow
	}
	if !exist && req.ActionType == -1 {
		return nil, dal.ErrNotFollow
	}

	// 检查关注数是否超过10k
	followCnt, err := dal.GetUserFollowCount(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	if followCnt >= 10000 {
		return nil, dal.ErrFollowLimit
	}

	// 通过kafka更新数据库
	err = kafka.Relation(ctx, &model.Relation{
		AuthorID:   req.ToUserId,
		FollowerID: req.UserId,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "通过kafka更新数据库失败")
		klog.Error("通过kafka更新数据库失败, err: ", err)
		return nil, err
	}

	// 返回响应
	resp = &relation.RelationActionResponse{}

	return
}

// RelationFollowList implements the RelationServiceImpl interface.
func (s *RelationServiceImpl) RelationFollowList(ctx context.Context, req *relation.RelationFollowListRequest) (resp *relation.RelationFollowListResponse, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "RelationFollowList")
	defer span.End()

	// 获取关注列表
	followList, err := dal.FollowList(ctx, req.ToUserId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "获取关注列表失败")
		klog.Error("获取关注列表失败, err: ", err)
		return nil, err
	}

	// 获取用户信息
	mUserList, err := dal.GetUserByIDs(ctx, followList)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "获取用户信息失败")
		klog.Error("获取用户信息失败, err: ", err)
		return nil, err
	}

	// 将model.User转换为user.User
	var wgUserList sync.WaitGroup
	wgUserList.Add(len(mUserList))
	userList := make([]*user.User, len(mUserList))
	for i, u := range mUserList {
		go func(i int, u *model.User) {
			defer wgUserList.Done()
			userList[i] = dal.ToUserResponse(ctx, req.UserId, u)
		}(i, u)
	}
	wgUserList.Wait()

	// 返回响应
	resp = &relation.RelationFollowListResponse{UserList: userList}

	return
}

// RelationFollowerList implements the RelationServiceImpl interface.
func (s *RelationServiceImpl) RelationFollowerList(ctx context.Context, req *relation.RelationFollowerListRequest) (resp *relation.RelationFollowerListResponse, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "RelationFollowerList")
	defer span.End()

	// 获取粉丝列表
	followerList, err := dal.FollowerList(ctx, req.ToUserId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "获取粉丝列表失败")
		klog.Error("获取粉丝列表失败, err: ", err)
		return nil, err
	}

	// 获取用户信息
	mUserList, err := dal.GetUserByIDs(ctx, followerList)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "获取用户信息失败")
		klog.Error("获取用户信息失败, err: ", err)
		return nil, err
	}

	// 将model.User转换为user.User
	var wgUserList sync.WaitGroup
	wgUserList.Add(len(mUserList))
	userList := make([]*user.User, len(mUserList))
	for i, u := range mUserList {
		go func(i int, u *model.User) {
			defer wgUserList.Done()
			userList[i] = dal.ToUserResponse(ctx, req.UserId, u)
		}(i, u)
	}
	wgUserList.Wait()

	// 返回响应
	resp = &relation.RelationFollowerListResponse{UserList: userList}

	return
}

// RelationFriendList implements the RelationServiceImpl interface.
func (s *RelationServiceImpl) RelationFriendList(ctx context.Context, req *relation.RelationFriendListRequest) (resp *relation.RelationFriendListResponse, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "RelationFriendList")
	defer span.End()

	// 获取好友列表
	friendList, err := dal.FriendList(ctx, req.ToUserId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "获取好友列表失败")
		klog.Error("获取好友列表失败, err: ", err)
		return nil, err
	}

	// 获取用户信息
	mFriendList, err := dal.GetUserByIDs(ctx, friendList)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "获取用户信息失败")
		klog.Error("获取用户信息失败, err: ", err)
		return nil, err
	}

	// 将model.User转换为user.User
	var wgFriendList sync.WaitGroup
	wgFriendList.Add(len(mFriendList))
	userList := make([]*user.User, len(mFriendList))
	for i, u := range mFriendList {
		go func(i int, u *model.User) {
			defer wgFriendList.Done()
			userList[i] = dal.ToUserResponse(ctx, req.UserId, u)
		}(i, u)
	}
	wgFriendList.Wait()

	// 返回响应
	resp = &relation.RelationFriendListResponse{UserList: userList}

	return
}
