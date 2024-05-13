package controller

import (
	"context"

	"douyin/src/dal"
	"douyin/src/pkg/jwt"
	"douyin/src/pkg/tracing"
	"douyin/src/rpc/client"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"go.opentelemetry.io/otel/codes"
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
	resp, err := client.RelationAction(c, userID, req.ToUserID, req.ActionType)
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
		if errorIs(err, dal.ErrFollowLimit){
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
	resp, err := client.FollowList(c, userID, req.UserID)
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
	resp, err := client.FollowerList(c, userID, req.UserID)
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
	resp, err := client.FriendList(c, userID, req.UserID)
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
