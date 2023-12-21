package controller

import (
	"context"
	"douyin/rpc/client"
	"io"
	"mime/multipart"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/kitex/pkg/klog"
)

type PublishController struct{}

type PublishActionRequest struct {
	Data  *multipart.FileHeader `form:"data"`                // 视频数据
	Title string                `form:"title" vd:"len($)>0"` // 视频标题
}

type PublishListRequest struct {
	UserID int64 `query:"user_id,string" vd:"$>0"` // 用户id
}

func NewPublishController() *PublishController {
	return &PublishController{}
}

func (pc *PublishController) Action(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &PublishActionRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
		klog.Error("PublishController.Action: 参数校验失败, err: ", err)
		return
	}

	// 验证大小
	if req.Data.Size > 1024*1024*100 {
		Error(ctx, CodeInvalidParam)
		klog.Error("PublishController.Action: 文件太大")
		return
	}

	// 从认证中间件中获取userID
	userID := ctx.MustGet(CtxUserIDKey).(int64)

	// 将文件转换为[]byte
	file, err := req.Data.Open()
	if err != nil {
		Error(ctx, CodeServerBusy)
		klog.Error("PublishController.Action: 文件打开失败, err: ", err)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		Error(ctx, CodeServerBusy)
		klog.Error("PublishController.Action: 文件读取失败, err: ", err)
		return
	}

	// 业务逻辑处理
	resp, err := client.PublishAction(userID, data, req.Title)
	if err != nil {
		Error(ctx, CodeServerBusy)
		klog.Error("PublishController.Action: 业务处理失败, err: ", err)
		return
	}

	// 返回响应
	Success(ctx, resp)
}

func (pc *PublishController) List(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &PublishListRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
		klog.Error("PublishController.List: 参数校验失败, err: ", err)
		return
	}
	authorID := req.UserID

	// 从认证中间件中获取userID
	userID := ctx.MustGet(CtxUserIDKey).(int64)

	// 业务逻辑处理
	resp, err := client.PublishList(userID, authorID)
	if err != nil {
		Error(ctx, CodeServerBusy)
		klog.Error("PublishController.Action: 业务处理失败, err: ", err)
		return
	}

	// 返回响应
	Success(ctx, resp)
}
