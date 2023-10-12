package controller

import (
	"context"
	"douyin/models"
	"douyin/response"
	"douyin/service"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func UserInfo(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &models.UserInfoRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		response.Error(ctx, response.CodeInvalidParam)
		hlog.Error("controller.UserInfo: 参数校验失败, err: ", err)
		return
	}

	// 业务逻辑处理
	resp, err := service.UserInfo(req)
	if err != nil {
		if errors.Is(err, service.ErrUserNotExist) {
			response.Error(ctx, response.CodeUserNotExist)
			hlog.Error("controller.UserInfo: 用户不存在")
			return
		}
		response.Error(ctx, response.CodeServerBusy)
		hlog.Error("controller.UserInfo: 业务处理失败, err: ", err)
		return
	}

	// 返回响应
	response.Success(ctx, resp)
}

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
		if errors.Is(err, service.ErrUserExist) {
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
	// 获取参数
	req := &models.UserRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		response.Error(ctx, response.CodeInvalidParam)
		hlog.Error("controller.Login: 参数校验失败, err: ", err)
		return
	}

	// 业务逻辑处理
	resp, err := service.Login(req)
	if err != nil {
		if errors.Is(err, service.ErrUserNotExist) {
			response.Error(ctx, response.CodeUserNotExist)
			hlog.Error("controller.Login: 用户不存在")
			return
		}
		if errors.Is(err, service.ErrPassword) {
			response.Error(ctx, response.CodeInvalidPassword)
			hlog.Error("controller.Login: 密码错误")
			return
		}
		response.Error(ctx, response.CodeServerBusy)
		hlog.Error("controller.Login: 业务处理失败, err: ", err)
		return
	}

	// 返回响应
	response.Success(ctx, resp)
}
