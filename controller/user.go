package controller

import (
	"context"
	"douyin/dao"
	"douyin/rpc/client"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/u2takey/go-utils/klog"
)

type UserController struct{}

type UserInfoRequest struct {
	UserID int64 `query:"user_id,string" vd:"$>0"` // 用户id
}

type UserRequest struct {
	Username string `query:"username" vd:"0<len($)&&len($)<33"` // 注册用户名，最长32个字符
	Password string `query:"password" vd:"5<len($)&&len($)<33"` // 密码，最长32个字符
}

func NewUserController() *UserController {
	return &UserController{}
}

func (uc *UserController) Info(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &UserInfoRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
		klog.Error("controller.UserInfo: 参数校验失败, err: ", err)
		return
	}

	// 从认证中间件中获取userID
	userID := ctx.MustGet(CtxUserIDKey).(int64)

	// 业务逻辑处理
	resp, err := client.UserInfo(req.UserID, userID)
	if err != nil {
		if errors.Is(err, dao.ErrUserNotExist) {
			Error(ctx, CodeUserNotExist)
			klog.Error("controller.UserInfo: 用户不存在")
			return
		}
		Error(ctx, CodeServerBusy)
		klog.Error("controller.UserInfo: 业务处理失败, err: ", err)
		return
	}

	// 返回响应
	Success(ctx, resp)
}

func (uc *UserController) Register(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &UserRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
		klog.Error("controller.Register: 参数校验失败, err: ", err)
		return
	}

	// 业务逻辑处理
	resp, err := client.Register(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, dao.ErrUserExist) {
			Error(ctx, CodeUserExist)
			klog.Error("controller.Register: 用户已存在")
			return
		}
		Error(ctx, CodeServerBusy)
		klog.Error("controller.Register: 业务处理失败, err: ", err)
		return
	}

	// 返回响应
	Success(ctx, resp)
}

func (uc *UserController) Login(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &UserRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
		klog.Error("controller.Login: 参数校验失败, err: ", err)
		return
	}

	// 业务逻辑处理
	resp, err := client.Login(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, dao.ErrUserNotExist) {
			Error(ctx, CodeUserNotExist)
			klog.Error("controller.Login: 用户不存在")
			return
		}
		if errors.Is(err, dao.ErrPassword) {
			Error(ctx, CodeInvalidPassword)
			klog.Error("controller.Login: 密码错误")
			return
		}
		Error(ctx, CodeServerBusy)
		klog.Error("controller.Login: 业务处理失败, err: ", err)
		return
	}

	// 返回响应
	Success(ctx, resp)
}
