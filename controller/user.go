package controller

import (
	"context"
	"douyin/dao/mysql"
	"douyin/models"
	"douyin/response"
	"douyin/service"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func Register(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &models.UserRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		response.Error(ctx, response.CodeInvalidParam)
		hlog.Error("controller.Register: 参数校验失败, err: ", err)
		return
	}

	// 业务逻辑处理
	resp, err := service.Register(req)
	if err != nil {
		if errors.Is(err, mysql.ErrUserExist) {
			response.Error(ctx, response.CodeUserExist)
			hlog.Error("controller.Register: 用户已存在")
			return
		}
		response.Error(ctx, response.CodeServerBusy)
		hlog.Error("controller.Register: 业务处理失败, err: ", err)
		return
	}

	// 返回响应
	response.Success(ctx, resp)
}

func Login(c context.Context, ctx *app.RequestContext) {
	hlog.Debug("controller.Login: 登录成功")
}

func UserInfo(c context.Context, ctx *app.RequestContext) {
	hlog.Debug("controller.UserInfo: 获取用户信息成功")
}
