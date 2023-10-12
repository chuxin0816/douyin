package controller

import (
	"context"
	"douyin/models"
	"douyin/response"
	"douyin/service"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func Register(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &models.UserRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		response.Error(ctx, response.CodeInvalidParam)
		hlog.Error("controller.Register: 参数校验失败", err)
		return
	}

	// 业务逻辑处理
	resp, err := service.Register(req)
	if err != nil {
		response.Error(ctx, response.CodeServerBusy)
		hlog.Error("controller.Register: 业务处理失败", err)
		return
	}

	// 返回响应
	response.Success(ctx, resp)
}
