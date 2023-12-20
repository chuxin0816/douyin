package controller

import (
	"context"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/u2takey/go-utils/klog"
)

type CommentController struct{}

type CommentActionRequest struct {
	VideoID     int64  `query:"video_id,string"    vd:"$>0"`                                            // 视频id
	ActionType  int64  `query:"action_type,string" vd:"$==1||$==2"`                                     // 1-发布评论，2-删除评论
	CommentID   int64  `query:"comment_id,string"  vd:"((ActionType)$==2&&$>0)||(ActionType)$==1"`      // 要删除的评论id，在action_type=2的时候使用
	CommentText string `query:"comment_text"       vd:"((ActionType)$==1&&len($)>0)||(ActionType)$==2"` // 用户填写的评论内容，在action_type=1的时候使用
}

type CommentListRequest struct {
	VideoID int64 `query:"video_id,string" vd:"$>0"` // 视频id
}

type CommentActionResponse struct {
	*Response
	Comment *CommentResponse `json:"comment,omitempty"`
}

type CommentListResponse struct {
	*Response
	CommentList []*CommentResponse `json:"comment_list"`
}

func NewCommentController() *CommentController {
	return &CommentController{}
}

func (cc *CommentController) Action(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &CommentActionRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
		klog.Error("CommentController.Action: 参数校验失败, err: ", err)
		return
	}

	// 从认证中间件中获取userID
	userID := ctx.MustGet(CtxUserIDKey).(int64)

	// 业务逻辑处理
	resp, err := service.CommentAction(userID, req.ActionType, req.VideoID, req.CommentID, req.CommentText)
	if err != nil {
		if errors.Is(err, dao.ErrVideoNotExist) {
			Error(ctx, CodeVideoNotExist)
			klog.Error("controller.CommentAction: 视频不存在")
			return
		}
		if errors.Is(err, dao.ErrCommentNotExist) {
			Error(ctx, CodeCommentNotExist)
			klog.Error("controller.CommentAction: 评论不存在")
			return
		}
		Error(ctx, CodeServerBusy)
		klog.Error("CommentController.Action: 业务逻辑处理失败, err: ", err)
		return
	}

	// 返回响应
	Success(ctx, resp)
}

func (cc *CommentController) List(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &CommentListRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
		klog.Error("CommentController.List: 参数校验失败, err: ", err)
		return
	}

	// 从认证中间件中获取userID
	userID := ctx.MustGet(CtxUserIDKey).(int64)

	// 业务逻辑处理
	resp, err := service.CommentList(userID, req.VideoID)
	if err != nil {
		if errors.Is(err, dao.ErrVideoNotExist) {
			Error(ctx, CodeVideoNotExist)
			klog.Error("controller.CommentList: 视频不存在")
			return
		}
		Error(ctx, CodeServerBusy)
		klog.Error("CommentController.List: 业务逻辑处理失败, err: ", err)
		return
	}

	// 返回响应
	Success(ctx, resp)
}
