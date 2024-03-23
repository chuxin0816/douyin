package main

import (
	"context"
	"douyin/config"
	"douyin/dal"
	"douyin/dal/model"
	relation "douyin/rpc/kitex_gen/relation"
	"douyin/rpc/kitex_gen/user"
	"strconv"

	"github.com/cloudwego/kitex/pkg/klog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

// RelationServiceImpl implements the last service interface defined in the IDL.
type RelationServiceImpl struct{}

// RelationAction implements the RelationServiceImpl interface.
func (s *RelationServiceImpl) RelationAction(ctx context.Context, req *relation.RelationActionRequest) (resp *relation.RelationActionResponse, err error) {
	ctx, span := otel.Tracer(config.Conf.OpenTelemetryConfig.RelationName).Start(ctx, "rpc.RelationAction")
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
		keyUserFollowCnt := dal.KeyUserFollowCountPF + strconv.FormatInt(req.UserId, 10)
		keyUserFollowerCnt := dal.KeyUserFollowerCountPF + strconv.FormatInt(req.ToUserId, 10)
		// 检查key是否存在
		if exist, err := dal.RDB.Exists(ctx, keyUserFollowCnt, keyUserFollowerCnt).Result(); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "检查key是否存在失败")
			klog.Error("检查key是否存在失败, err: ", err)
			return
		} else if exist != 2 {
			// 缓存不存在，查询数据库写入缓存
			cnt, err := dal.GetUserFollowCount(ctx, req.UserId)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "查询数据库失败")
				klog.Error("查询数据库失败, err: ", err)
				return
			}
			if err = dal.RDB.Set(ctx, keyUserFollowCnt, cnt, 0).Err(); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "写入缓存失败")
				klog.Error("写入缓存失败, err: ", err)
				return
			}
			cnt, err = dal.GetUserFollowerCount(ctx, req.ToUserId)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "查询数据库失败")
				klog.Error("查询数据库失败, err: ", err)
				return
			}
			if err = dal.RDB.Set(ctx, keyUserFollowerCnt, cnt, 0).Err(); err != nil {
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
		dal.CacheUserID.Store(req.UserId, struct{}{})
		dal.CacheUserID.Store(req.ToUserId, struct{}{})
	}()

	// 返回响应
	resp = &relation.RelationActionResponse{}

	return
}

// RelationFollowList implements the RelationServiceImpl interface.
func (s *RelationServiceImpl) RelationFollowList(ctx context.Context, req *relation.RelationFollowListRequest) (resp *relation.RelationFollowListResponse, err error) {
	ctx, span := otel.Tracer(config.Conf.OpenTelemetryConfig.RelationName).Start(ctx, "rpc.RelationFollowList")
	defer span.End()

	// 操作数据库
	mUserList, err := dal.FollowList(ctx, req.ToUserId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "操作数据库失败")
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
	ctx, span := otel.Tracer(config.Conf.OpenTelemetryConfig.RelationName).Start(ctx, "rpc.RelationFollowerList")
	defer span.End()

	// 操作数据库
	mUserList, err := dal.FollowerList(ctx, req.ToUserId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "操作数据库失败")
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
	ctx, span := otel.Tracer(config.Conf.OpenTelemetryConfig.RelationName).Start(ctx, "rpc.RelationFriendList")
	defer span.End()

	// 获取关注列表
	mFollowList, err := dal.FollowList(ctx, req.ToUserId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "获取关注列表失败")
		klog.Error("获取关注列表失败, err: ", err)
		return nil, err
	}

	// 获取粉丝列表
	mFollowerList, err := dal.FollowerList(ctx, req.ToUserId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "获取粉丝列表失败")
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
