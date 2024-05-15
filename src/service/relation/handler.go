package main

import (
	"context"
	"strconv"
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

	// 添加缓存避免重复操作
	keyUserFollow := dal.GetRedisKey(dal.KeyUserFollowPF + strconv.FormatInt(req.UserId, 10))
	if req.ActionType == 1 {
		dal.RDB.SAdd(ctx, keyUserFollow, req.ToUserId)
		dal.RDB.Expire(ctx, keyUserFollow, dal.ExpireTime+dal.GetRandomTime())
	} else {
		dal.RDB.SRem(ctx, keyUserFollow, req.ToUserId)
	}

	// 通过kafka更新数据库
	err = kafka.Relation(ctx, &model.Relation{
		AuthorID:   req.ToUserId,
		FollowerID: req.UserId,
	})
	if err != nil {
		// 更新失败，回滚缓存
		if req.ActionType == 1 {
			dal.RDB.SRem(ctx, keyUserFollow, req.ToUserId)
		} else {
			dal.RDB.SAdd(ctx, keyUserFollow, req.ToUserId)
			dal.RDB.Expire(ctx, keyUserFollow, dal.ExpireTime+dal.GetRandomTime())
		}

		span.RecordError(err)
		span.SetStatus(codes.Error, "通过kafka更新数据库失败")
		klog.Error("通过kafka更新数据库失败, err: ", err)
		return nil, err
	}

	// 更新缓存相关字段
	keyUserFollowCnt := dal.GetRedisKey(dal.KeyUserFollowCountPF + strconv.FormatInt(req.UserId, 10))
	keyUserFollowerCnt := dal.GetRedisKey(dal.KeyUserFollowerCountPF + strconv.FormatInt(req.ToUserId, 10))
	// 检查key是否存在
	if exist, err := dal.RDB.Exists(ctx, keyUserFollowerCnt).Result(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "检查key是否存在失败")
		klog.Error("检查key是否存在失败, err: ", err)
		return nil, err
	} else if exist == 0 {
		// 缓存不存在，查询数据库写入缓存
		cnt, err := dal.GetUserFollowerCount(ctx, req.ToUserId)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "查询数据库失败")
			klog.Error("查询数据库失败, err: ", err)
			return nil, err
		}
		if err = dal.RDB.Set(ctx, keyUserFollowerCnt, cnt, 0).Err(); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "写入缓存失败")
			klog.Error("写入缓存失败, err: ", err)
			return nil, err
		}
	}

	pipe := dal.RDB.Pipeline()
	pipe.IncrBy(ctx, keyUserFollowCnt, req.ActionType)
	pipe.IncrBy(ctx, keyUserFollowerCnt, req.ActionType)
	if _, err := pipe.Exec(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "更新缓存失败")
		klog.Error("更新缓存失败, err: ", err)
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

	// 获取关注列表
	followList, err := dal.FollowList(ctx, req.ToUserId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "获取关注列表失败")
		klog.Error("获取关注列表失败, err: ", err)
		return nil, err
	}

	// 获取好友列表
	friendList := make([]int64, 0, len(followList))
	for _, id := range followList {
		exist, err := dal.CheckRelationExist(ctx, id, req.ToUserId)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "检查是否关注失败")
			klog.Error("检查是否关注失败, err: ", err)
			return nil, err
		}
		if exist {
			friendList = append(friendList, id)
		}
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
