package controller

import (
	"context"

	"douyin/src/dal"
	"douyin/src/pkg/jwt"
	"douyin/src/pkg/tracing"

	"douyin/src/config"
	"douyin/src/kitex_gen/relation"
	"douyin/src/kitex_gen/relation/relationservice"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"go.opentelemetry.io/otel/codes"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	tracing2 "github.com/kitex-contrib/obs-opentelemetry/tracing"
	consul "github.com/kitex-contrib/registry-consul"
)

type RelationController struct{}

type RelationActionRequest struct {
	ToUserID   int64 `query:"to_user_id,string"  vd:"$>0"`        // 对方用户id
	ActionType int64 `query:"action_type,string" vd:"$==1||$==2"` // 1-关注，2-取消关注
}

type RelationListRequest struct {
	UserID int64  `query:"user_id,string" vd:"$>0"` // 用户id
	Token  string `query:"token"`                   // 用户登录状态下设置
}

var relationClient relationservice.Client

func init() {
	// 服务发现
	r, err := consul.NewConsulResolver(config.Conf.ConsulConfig.ConsulAddr)
	if err != nil {
		panic(err)
	}

	relationClient, err = relationservice.NewClient(
		config.Conf.OpenTelemetryConfig.RelationName,
		client.WithResolver(r),
		client.WithSuite(tracing2.NewClientSuite()),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: config.Conf.OpenTelemetryConfig.ApiName}),
		client.WithMuxConnection(2),
	)
	if err != nil {
		panic(err)
	}
}

func NewRelationController() *RelationController {
	return &RelationController{}
}

func (rc *RelationController) Action(c context.Context, ctx *app.RequestContext) {
	c, span := tracing.Tracer.Start(c, "RelationAction")
	defer span.End()

	// 获取参数
	req := &RelationActionRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
		span.RecordError(err)
		span.SetStatus(codes.Error, "参数校验失败")
		hlog.Error("参数校验失败, err: ", err)
		return
	}

	// 解析关注类型
	if req.ActionType == 2 {
		req.ActionType = -1
	}

	// 从认证中间件中获取userID
	userID := ctx.MustGet(CtxUserIDKey).(int64)

	// 业务逻辑处理
	resp, err := relationClient.RelationAction(c, &relation.RelationActionRequest{
		UserId:     userID,
		ToUserId:   req.ToUserID,
		ActionType: req.ActionType,
	})
	if err != nil {
		span.RecordError(err)
		if errorIs(err, dal.ErrAlreadyFollow) {
			Error(ctx, CodeAlreadyFollow)
			span.SetStatus(codes.Error, "已经关注过了")
			hlog.Error("已经关注过了, err: ", err)
			return
		}
		if errorIs(err, dal.ErrNotFollow) {
			Error(ctx, CodeNotFollow)
			span.SetStatus(codes.Error, "还没有关注过")
			hlog.Error("还没有关注过, err: ", err)
			return
		}
		if errorIs(err, dal.ErrFollowLimit) {
			Error(ctx, CodeFollowLimit)
			span.SetStatus(codes.Error, "关注数超过上限")
			hlog.Error("关注数超过上限, err: ", err)
			return
		}
		Error(ctx, CodeServerBusy)
		span.SetStatus(codes.Error, "业务逻辑处理失败")
		hlog.Error("业务逻辑处理失败, err: ", err)
		return
	}

	// 返回响应
	Success(ctx, resp)
}

func (rc *RelationController) FollowList(c context.Context, ctx *app.RequestContext) {
	c, span := tracing.Tracer.Start(c, "RelationFollowList")
	defer span.End()

	// 获取参数
	req := &RelationListRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
		span.RecordError(err)
		span.SetStatus(codes.Error, "参数校验失败")
		hlog.Error("参数校验失败, err: ", err)
		return
	}

	// 验证token
	userID := jwt.ParseToken(req.Token)

	// 业务逻辑处理
	resp, err := relationClient.RelationFollowList(c, &relation.RelationFollowListRequest{
		UserId:   userID,
		ToUserId: req.UserID,
	})
	if err != nil {
		Error(ctx, CodeServerBusy)
		span.RecordError(err)
		span.SetStatus(codes.Error, "业务逻辑处理失败")
		hlog.Error("业务逻辑处理失败, err: ", err)
		return
	}

	// 返回响应
	Success(ctx, resp)
}

func (rc *RelationController) FollowerList(c context.Context, ctx *app.RequestContext) {
	c, span := tracing.Tracer.Start(c, "RelationFollowerList")
	defer span.End()

	// 获取参数
	req := &RelationListRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
		span.RecordError(err)
		span.SetStatus(codes.Error, "参数校验失败")
		hlog.Error("参数校验失败, err: ", err)
		return
	}

	// 验证token
	userID := jwt.ParseToken(req.Token)

	// 业务逻辑处理
	resp, err := relationClient.RelationFollowerList(c, &relation.RelationFollowerListRequest{
		UserId:   userID,
		ToUserId: req.UserID,
	})
	if err != nil {
		Error(ctx, CodeServerBusy)
		span.RecordError(err)
		span.SetStatus(codes.Error, "业务逻辑处理失败")
		hlog.Error("业务逻辑处理失败, err: ", err)
		return
	}

	// 返回响应
	Success(ctx, resp)
}

func (rc *RelationController) FriendList(c context.Context, ctx *app.RequestContext) {
	c, span := tracing.Tracer.Start(c, "RelationFriendList")
	defer span.End()

	// 获取参数
	req := &RelationListRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
		span.RecordError(err)
		span.SetStatus(codes.Error, "参数校验失败")
		hlog.Error("参数校验失败, err: ", err)
		return
	}

	// 验证token
	userID := jwt.ParseToken(req.Token)

	// 业务逻辑处理
	resp, err := relationClient.RelationFriendList(c, &relation.RelationFriendListRequest{
		UserId:   userID,
		ToUserId: req.UserID,
	})
	if err != nil {
		Error(ctx, CodeServerBusy)
		span.RecordError(err)
		span.SetStatus(codes.Error, "业务逻辑处理失败")
		hlog.Error("业务逻辑处理失败, err: ", err)
		return
	}

	// 返回响应
	Success(ctx, resp)
}
