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

type FavoriteController struct{}

type FavoriteActionRequest struct {
	VideoID    int64 `query:"video_id,string"    vd:"$>0"`        // 视频id
	ActionType int64 `query:"action_type,string" vd:"$==1||$==2"` // 1-点赞，2-取消点赞
}

type FavoriteListRequest struct {
	UserID int64  `query:"user_id,string" vd:"$>0"` // 用户id
	Token  string `query:"token"`                   // 用户登录状态下设置
}

func NewFavoriteController() *FavoriteController {
	return &FavoriteController{}
}

func (fc *FavoriteController) Action(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &FavoriteActionRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
		klog.Error("参数校验失败, err: ", err)
		return
	}

	// 从认证中间件中获取userID
	userID := ctx.MustGet(CtxUserIDKey).(int64)

	// 业务逻辑处理
	resp, err := client.FavoriteAction(userID, req.VideoID, req.ActionType)
	if err != nil {
		if errors.Is(err, dal.ErrAlreadyFavorite) {
			Error(ctx, CodeAlreadyFavorite)
			klog.Error("已经点赞过了")
			return
		}
		if errors.Is(err, dal.ErrNotFavorite) {
			Error(ctx, CodeNotFavorite)
			klog.Error("还没有点赞过")
			return
		}
		Error(ctx, CodeServerBusy)
		klog.Error("业务处理失败, err: ", err)
		return
	}

	// 返回响应
	Success(ctx, resp)
}

func (fc *FavoriteController) List(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &FavoriteListRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
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
	resp, err := client.FavoriteList(userID, req.UserID)
	if err != nil {
		Error(ctx, CodeServerBusy)
		klog.Error("业务处理失败, err: ", err)
		return
	}

	// 返回响应
	Success(ctx, resp)
}