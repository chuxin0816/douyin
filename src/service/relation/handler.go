package main

import (
	"context"

	"douyin/src/config"
	"douyin/src/dal"
	"douyin/src/dal/model"
	relation "douyin/src/kitex_gen/relation"
	"douyin/src/kitex_gen/user"
	"douyin/src/kitex_gen/user/userservice"
	"douyin/src/pkg/kafka"
	"douyin/src/pkg/tracing"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	tracing2 "github.com/kitex-contrib/obs-opentelemetry/tracing"
	consul "github.com/kitex-contrib/registry-consul"
	"go.opentelemetry.io/otel/codes"
)

// RelationServiceImpl implements the last service interface defined in the IDL.
type RelationServiceImpl struct{}

var userClient userservice.Client

func init() {
	// 服务发现
	r, err := consul.NewConsulResolver(config.Conf.ConsulConfig.ConsulAddr)
	if err != nil {
		panic(err)
	}

	userClient, err = userservice.NewClient(
		config.Conf.OpenTelemetryConfig.UserName,
		client.WithResolver(r),
		client.WithSuite(tracing2.NewClientSuite()),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: config.Conf.OpenTelemetryConfig.UserName}),
		client.WithMuxConnection(2),
	)
	if err != nil {
		panic(err)
	}
}

// RelationAction implements the RelationServiceImpl interface.
func (s *RelationServiceImpl) RelationAction(ctx context.Context, req *relation.RelationActionRequest) (resp *relation.RelationActionResponse, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "RelationAction")
	defer span.End()

	// 检查是否关注
	exist, err := dal.CheckRelationExist(ctx, req.UserId, req.AuthorId)
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
		AuthorID:   req.AuthorId,
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
	followList, err := dal.FollowList(ctx, req.AuthorId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "获取关注列表失败")
		klog.Error("获取关注列表失败, err: ", err)
		return nil, err
	}

	// 获取用户信息
	userList := make([]*user.User, len(followList))
	for i, u := range followList {
		user, err := userClient.UserInfo(ctx, &user.UserInfoRequest{
			UserId:   req.UserId,
			AuthorId: u,
		})
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "获取用户信息失败")
			klog.Error("获取用户信息失败, err: ", err)
			return nil, err
		}
		userList[i] = user.User
	}

	// 返回响应
	resp = &relation.RelationFollowListResponse{UserList: userList}

	return
}

// RelationFollowerList implements the RelationServiceImpl interface.
func (s *RelationServiceImpl) RelationFollowerList(ctx context.Context, req *relation.RelationFollowerListRequest) (resp *relation.RelationFollowerListResponse, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "RelationFollowerList")
	defer span.End()

	// 获取粉丝列表
	followerList, err := dal.FollowerList(ctx, req.AuthorId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "获取粉丝列表失败")
		klog.Error("获取粉丝列表失败, err: ", err)
		return nil, err
	}

	// 获取用户信息
	// 获取用户信息
	userList := make([]*user.User, len(followerList))
	for i, u := range followerList {
		user, err := userClient.UserInfo(ctx, &user.UserInfoRequest{
			UserId:   req.UserId,
			AuthorId: u,
		})
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "获取用户信息失败")
			klog.Error("获取用户信息失败, err: ", err)
			return nil, err
		}
		userList[i] = user.User
	}

	// 返回响应
	resp = &relation.RelationFollowerListResponse{UserList: userList}

	return
}

// RelationFriendList implements the RelationServiceImpl interface.
func (s *RelationServiceImpl) RelationFriendList(ctx context.Context, req *relation.RelationFriendListRequest) (resp *relation.RelationFriendListResponse, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "RelationFriendList")
	defer span.End()

	// 获取好友列表
	friendList, err := dal.FriendList(ctx, req.AuthorId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "获取好友列表失败")
		klog.Error("获取好友列表失败, err: ", err)
		return nil, err
	}

	// 获取用户信息
	userList := make([]*user.User, len(friendList))
	for i, u := range friendList {
		user, err := userClient.UserInfo(ctx, &user.UserInfoRequest{
			UserId:   req.UserId,
			AuthorId: u,
		})
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "获取用户信息失败")
			klog.Error("获取用户信息失败, err: ", err)
			return nil, err
		}
		userList[i] = user.User
	}

	// 返回响应
	resp = &relation.RelationFriendListResponse{UserList: userList}

	return
}

// RelationExist implements the RelationServiceImpl interface.
func (s *RelationServiceImpl) RelationExist(ctx context.Context, userId int64, authorId int64) (resp bool, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "RelationExist")
	defer span.End()

	resp, err = dal.CheckRelationExist(ctx, userId, authorId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "检查关注关系失败")
		klog.Error("检查关注关系失败, err: ", err)
		return
	}

	return
}

// FollowCnt implements the RelationServiceImpl interface.
func (s *RelationServiceImpl) FollowCnt(ctx context.Context, userId int64) (resp int64, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "FollowCnt")
	defer span.End()

	resp, err = dal.GetUserFollowCount(ctx, userId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "获取关注数失败")
		klog.Error("获取关注数失败, err: ", err)
		return
	}

	return
}

// FollowerCnt implements the RelationServiceImpl interface.
func (s *RelationServiceImpl) FollowerCnt(ctx context.Context, userId int64) (resp int64, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "FollowerCnt")
	defer span.End()

	resp, err = dal.GetUserFollowerCount(ctx, userId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "获取粉丝数失败")
		klog.Error("获取粉丝数失败, err: ", err)
		return
	}

	return
}
