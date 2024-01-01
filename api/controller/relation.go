package controller

import (
	"context"
	"douyin/dal"
	"douyin/pkg/jwt"
	"douyin/rpc/client"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/kitex/pkg/klog"
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
	// 获取参数
	req := &RelationActionRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		klog.Error("参数校验失败, err: ", err)
		Error(ctx, CodeInvalidParam)
		return
	}

	// 从认证中间件中获取userID
	userID := ctx.MustGet(CtxUserIDKey).(int64)

	// 业务逻辑处理
	resp, err := client.RelationAction(userID, req.ToUserID, req.ActionType)
	if err != nil {
		if errors.Is(err, dal.ErrAlreadyFollow) {
			klog.Error("已经关注过了, err: ", err)
			Error(ctx, CodeAlreadyFollow)
			return
		}
		if errors.Is(err, dal.ErrNotFollow) {
			klog.Error("还没有关注过, err: ", err)
			Error(ctx, CodeNotFollow)
			return
		}
		klog.Error("业务逻辑处理失败, err: ", err)
		Error(ctx, CodeServerBusy)
		return
	}

	// 返回响应
	Success(ctx, resp)
}

func (rc *RelationController) FollowList(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &RelationListRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		klog.Error("参数校验失败, err: ", err)
		Error(ctx, CodeInvalidParam)
		return
	}

	// 验证token
	var userID *int64
	if len(req.Token) > 0 {
		userID = jwt.ParseToken(req.Token)
		if userID == nil {
			Error(ctx, CodeNoAuthority)
			klog.Error("token解析失败")
			return
		}
	}

	// 业务逻辑处理
	resp, err := client.FollowList(userID, req.UserID)
	if err != nil {
		klog.Error("业务逻辑处理失败, err: ", err)
		Error(ctx, CodeServerBusy)
		return
	}

	// 返回响应
	Success(ctx, resp)
}

func (rc *RelationController) FollowerList(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &RelationListRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		klog.Error("参数校验失败, err: ", err)
		return
	}

	// 验证token
	var userID *int64
	if len(req.Token) > 0 {
		userID = jwt.ParseToken(req.Token)
		if userID == nil {
			Error(ctx, CodeNoAuthority)
			klog.Error("token解析失败")
			return
		}
	}

	// 业务逻辑处理
	resp, err := client.FollowerList(userID, req.UserID)
	if err != nil {
		klog.Error("业务逻辑处理失败, err: ", err)
		Error(ctx, CodeServerBusy)
		return
	}

	// 返回响应
	Success(ctx, resp)
}

func (rc *RelationController) FriendList(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &RelationListRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		klog.Error("参数校验失败, err: ", err)
		return
	}

	// 验证token
	var userID *int64
	if len(req.Token) > 0 {
		userID = jwt.ParseToken(req.Token)
		if userID == nil {
			Error(ctx, CodeNoAuthority)
			klog.Error("token解析失败")
			return
		}
	}

	// 业务逻辑处理
	resp, err := client.FriendList(userID, req.UserID)
	if err != nil {
		klog.Error("业务逻辑处理失败, err: ", err)
		Error(ctx, CodeServerBusy)
		return
	}

	// 返回响应
	Success(ctx, resp)
}