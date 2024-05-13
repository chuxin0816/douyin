package controller

import (
	"context"
	"io"
	"mime/multipart"

	"douyin/src/pkg/jwt"
	"douyin/src/pkg/tracing"
	"douyin/src/rpc/client"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"go.opentelemetry.io/otel/codes"
)

const(
	minFileSize = 1 * 1024 * 1024 // 1MB
	maxFileSize = 500 * 1024 * 1024 // 500MB
)

type PublishController struct{}

type PublishActionRequest struct {
	Data  *multipart.FileHeader `form:"data"`                // 视频数据
	Title string                `form:"title" vd:"len($)>0"` // 视频标题
}

type PublishListRequest struct {
	UserID int64  `query:"user_id,string" vd:"$>0"` // 用户id
	Token  string `query:"token"`                   // 用户登录状态下设置
}

func NewPublishController() *PublishController {
	return &PublishController{}
}

func (pc *PublishController) Action(c context.Context, ctx *app.RequestContext) {
	c, span := tracing.Tracer.Start(c, "PublishAction")
	defer span.End()

	// 获取参数
	req := &PublishActionRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
		span.RecordError(err)
		span.SetStatus(codes.Error, "参数校验失败")
		hlog.Error("参数校验失败, err: ", err)
		return
	}

	// 检查标题字数
	if len(req.Title) > 30 {
		Error(ctx, codeLengthLimit)
		hlog.Warn("标题字数超过限制")
		return
	}

	// 验证大小
	if req.Data.Size < minFileSize {
		Error(ctx, CodeFileTooSmall)
		hlog.Warn("文件太小")
		return
	}
	if req.Data.Size > maxFileSize {
		Error(ctx, CodeFileTooLarge)
		hlog.Warn("文件太大")
		return
	}

	// 从认证中间件中获取userID
	userID := ctx.MustGet(CtxUserIDKey).(int64)

	// 将文件转换为[]byte
	file, err := req.Data.Open()
	if err != nil {
		Error(ctx, CodeServerBusy)
		span.RecordError(err)
		span.SetStatus(codes.Error, "文件打开失败")
		hlog.Error("文件打开失败, err: ", err)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		Error(ctx, CodeServerBusy)
		span.RecordError(err)
		span.SetStatus(codes.Error, "文件读取失败")
		hlog.Error("文件读取失败, err: ", err)
		return
	}

	// 业务逻辑处理
	resp, err := client.PublishAction(c, userID, data, req.Title)
	if err != nil {
		Error(ctx, CodeServerBusy)
		span.RecordError(err)
		span.SetStatus(codes.Error, "业务处理失败")
		hlog.Error("业务处理失败, err: ", err)
		return
	}

	// 返回响应
	Success(ctx, resp)
}

func (pc *PublishController) List(c context.Context, ctx *app.RequestContext) {
	c, span := tracing.Tracer.Start(c, "PublishList")
	defer span.End()

	// 获取参数
	req := &PublishListRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
		span.RecordError(err)
		span.SetStatus(codes.Error, "参数校验失败")
		hlog.Error("参数校验失败, err: ", err)
		return
	}
	authorID := req.UserID

	// 验证token
	userID := jwt.ParseToken(req.Token)

	// 业务逻辑处理
	resp, err := client.PublishList(c, userID, authorID)
	if err != nil {
		Error(ctx, CodeServerBusy)
		span.RecordError(err)
		span.SetStatus(codes.Error, "业务处理失败")
		hlog.Error("业务处理失败, err: ", err)
		return
	}

	// 返回响应
	Success(ctx, resp)
}
