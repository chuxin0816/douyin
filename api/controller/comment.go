package controller

import (
	"context"
	"douyin/dal"
	"douyin/pkg/jwt"
	"douyin/pkg/tracing"
	"douyin/rpc/client"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/kitex/pkg/klog"

	"go.opentelemetry.io/otel/codes"
)

type CommentController struct{}

type CommentActionRequest struct {
	VideoID     int64  `query:"video_id,string"    vd:"$>0"`                                            // 视频id
	ActionType  int64  `query:"action_type,string" vd:"$==1||$==2"`                                     // 1-发布评论，2-删除评论
	CommentID   int64  `query:"comment_id,string"  vd:"((ActionType)$==2&&$>0)||(ActionType)$==1"`      // 要删除的评论id，在action_type=2的时候使用
	CommentText string `query:"comment_text"       vd:"((ActionType)$==1&&len($)>0)||(ActionType)$==2"` // 用户填写的评论内容，在action_type=1的时候使用
}

type CommentListRequest struct {
	VideoID int64  `query:"video_id,string" vd:"$>0"` // 视频id
	Token   string `query:"token"`                    // 用户登录状态下设置
}

func NewCommentController() *CommentController {
	return &CommentController{}
}

func (cc *CommentController) Action(c context.Context, ctx *app.RequestContext) {
	_, span := tracing.Tracer.Start(c, "CommentAction")
	defer span.End()
	// 获取参数
	req := &CommentActionRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
		span.RecordError(err)
		span.SetStatus(codes.Error, "参数校验失败")
		klog.Error("参数校验失败, err: ", err)
		return
	}

	// 从认证中间件中获取userID
	userID := ctx.MustGet(CtxUserIDKey).(int64)

	// 业务逻辑处理
	resp, err := client.CommentAction(userID, req.ActionType, req.VideoID, &req.CommentID, &req.CommentText)
	if err != nil {
		span.RecordError(err)

		if errorIs(err, dal.ErrVideoNotExist) {
			Error(ctx, CodeVideoNotExist)
			span.SetStatus(codes.Error, "视频不存在")
			klog.Error("视频不存在")
			return
		}
		if errorIs(err, dal.ErrCommentNotExist) {
			Error(ctx, CodeCommentNotExist)
			span.SetStatus(codes.Error, "评论不存在")
			klog.Error("评论不存在")
			return
		}
		Error(ctx, CodeServerBusy)
		span.SetStatus(codes.Error, "业务逻辑处理失败")
		klog.Error("业务逻辑处理失败, err: ", err)
		return
	}

	// 返回响应
	Success(ctx, resp)
}

func (cc *CommentController) List(c context.Context, ctx *app.RequestContext) {
	_, span := tracing.Tracer.Start(c, "CommentList")
	defer span.End()

	// 获取参数
	req := &CommentListRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
		span.RecordError(err)
		span.SetStatus(codes.Error, "参数校验失败")
		klog.Error("参数校验失败, err: ", err)
		return
	}

	// 验证token
	userID := jwt.ParseToken(req.Token)

	// 业务逻辑处理
	resp, err := client.CommentList(userID, req.VideoID)
	if err != nil {
		span.RecordError(err)
		if errorIs(err, dal.ErrVideoNotExist) {
			Error(ctx, CodeVideoNotExist)
			span.SetStatus(codes.Error, "视频不存在")
			klog.Error("视频不存在")
			return
		}
		Error(ctx, CodeServerBusy)
		span.SetStatus(codes.Error, "业务逻辑处理失败")
		klog.Error("业务逻辑处理失败, err: ", err)
		return
	}

	// 返回响应
	Success(ctx, resp)
}
