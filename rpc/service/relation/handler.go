package main

import (
	"context"
	"strconv"
	"sync"

	"douyin/dal"
	"douyin/dal/model"
	"douyin/pkg/tracing"
	relation "douyin/rpc/kitex_gen/relation"
	"douyin/rpc/kitex_gen/user"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/redis/go-redis/v9"
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
	var followCnt int64
	keyUserFollowCnt := dal.GetRedisKey(dal.KeyUserFollowCountPF + strconv.FormatInt(req.UserId, 10))
	cnt, err := dal.RDB.Get(ctx, keyUserFollowCnt).Result()
	if err == redis.Nil {
		followCnt, err = dal.GetUserFollowCount(ctx, req.UserId)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "查询数据库失败")
			klog.Error("查询数据库失败, err: ", err)
			return nil, err
		}
		if err = dal.RDB.Set(ctx, keyUserFollowCnt, followCnt, redis.KeepTTL).Err(); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "写入缓存失败")
			klog.Error("写入缓存失败, err: ", err)
			return nil, err
		}
	} else if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "查询缓存失败")
		klog.Error("查询缓存失败, err: ", err)
		return nil, err
	} else {
		followCnt, err = strconv.ParseInt(cnt, 10, 64)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "类型转换失败")
			klog.Error("类型转换失败, err: ", err)
			return nil, err
		}
	}

	if followCnt >= 10000 {
		return nil, dal.ErrFollowLimit
	}

	// 操作数据库
	if req.ActionType == 1 {
		if err := dal.Follow(ctx, req.UserId, req.ToUserId); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "关注失败")
			klog.Error("关注失败, err: ", err)
			return nil, err
		}
	} else {
		if err := dal.UnFollow(ctx, req.UserId, req.ToUserId); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "取消关注失败")
			klog.Error("取消关注失败, err: ", err)
			return nil, err
		}
	}

	// 更新缓存相关字段
	go func() {
		keyUserFollowerCnt := dal.GetRedisKey(dal.KeyUserFollowerCountPF + strconv.FormatInt(req.ToUserId, 10))
		// 检查key是否存在
		if exist, err := dal.RDB.Exists(ctx, keyUserFollowerCnt).Result(); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "检查key是否存在失败")
			klog.Error("检查key是否存在失败, err: ", err)
			return
		} else if exist == 0 {
			// 缓存不存在，查询数据库写入缓存
			cnt, err := dal.GetUserFollowerCount(ctx, req.ToUserId)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "查询数据库失败")
				klog.Error("查询数据库失败, err: ", err)
				return
			}
			if err = dal.RDB.Set(ctx, keyUserFollowerCnt, cnt, redis.KeepTTL).Err(); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "写入缓存失败")
				klog.Error("写入缓存失败, err: ", err)
				return
			}
		}

		if err := dal.RDB.IncrBy(ctx, keyUserFollowCnt, req.ActionType).Err(); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "更新user的follow_count字段失败")
			klog.Error("更新user的follow_count字段失败, err: ", err)
			return
		}
		if err := dal.RDB.IncrBy(ctx, keyUserFollowerCnt, req.ActionType).Err(); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "更新user的follower_count字段失败")
			klog.Error("更新user的follower_count字段失败, err: ", err)
			return
		}
		// 写入待同步切片
		dal.Mu.Lock()
		dal.CacheUserID[req.UserId] = struct{}{}
		dal.CacheUserID[req.ToUserId] = struct{}{}
		dal.Mu.Unlock()
	}()

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

	// 获取粉丝列表
	followerList, err := dal.FollowerList(ctx, req.ToUserId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "获取粉丝列表失败")
		klog.Error("获取粉丝列表失败, err: ", err)
		return nil, err
	}

	// 获取好友列表
	size := min(len(followList), len(followerList))
	friendList := make([]int64, 0, size)
	mp := make(map[int64]struct{}, len(followList))
	for _, userID := range followList {
		mp[userID] = struct{}{}
	}

	for _, userID := range followerList {
		if _, ok := mp[userID]; ok {
			friendList = append(friendList, userID)
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
