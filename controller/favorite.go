package controller

import (
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/u2takey/go-utils/klog"
)

type FavoriteController struct{}

type FavoriteActionRequest struct {
	VideoID    int64 `query:"video_id,string"    vd:"$>0"`        // 视频id
	ActionType int64 `query:"action_type,string" vd:"$==1||$==2"` // 1-点赞，2-取消点赞
}

type FavoriteListRequest struct {
	UserID int64 `query:"user_id,string" vd:"$>0"` // 用户id
}

type FavoriteActionResponse struct {
	*Response
}

type FavoriteListResponse struct {
	*Response
	VideoList []*VideoResponse `json:"video_list"`
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
		klog.Error("FavoriteController: 参数校验失败, err: ", err)
		return
	}

	// 从认证中间件中获取userID
	userID := ctx.MustGet(CtxUserIDKey).(int64)

	// 业务逻辑处理
	resp, err := service.FavoriteAction(userID, req.VideoID, req.ActionType)
	if err != nil {
		if errors.Is(err, dao.ErrAlreadyFavorite) {
			Error(ctx, CodeAlreadyFavorite)
			klog.Error("FavoriteController.Action: 已经点赞过了")
			return
		}
		if errors.Is(err, dao.ErrNotFavorite) {
			Error(ctx, CodeNotFavorite)
			klog.Error("FavoriteController.Action: 还没有点赞过")
			return
		}
		Error(ctx, CodeServerBusy)
		klog.Error("FavoriteController.Action: 业务处理失败, err: ", err)
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
		klog.Error("FavoriteController: 参数校验失败, err: ", err)
		return
	}

	// 从认证中间件中获取userID
	userID := ctx.MustGet(CtxUserIDKey).(int64)

	// 业务逻辑处理
	resp, err := service.FavoriteList(userID, req.UserID)
	if err != nil {
		Error(ctx, CodeServerBusy)
		klog.Error("FavoriteController.List: 业务处理失败, err: ", err)
		return
	}

	// 返回响应
	Success(ctx, resp)
}
