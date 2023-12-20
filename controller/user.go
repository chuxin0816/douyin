package controller

import (
	"context"
	"douyin/dao"
	"douyin/middleware"
	"douyin/response"
	"douyin/service"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type UserController struct{}

type UserInfoRequest struct {
	UserID int64 `query:"user_id,string" vd:"$>0"` // 用户id
}

type UserRequest struct {
	Username string `query:"username" vd:"0<len($)&&len($)<33"` // 注册用户名，最长32个字符
	Password string `query:"password" vd:"5<len($)&&len($)<33"` // 密码，最长32个字符
}

type UserInfoResponse struct {
	*Response
	User *UserResponse `json:"user"` // 用户信息
}

type RegisterResponse struct {
	*Response
	UserID int64  `json:"user_id"` // 用户id
	Token  string `json:"token"`   // 用户鉴权token
}

type LoginResponse struct {
	*Response
	UserID int64  `json:"user_id"` // 用户id
	Token  string `json:"token"`   // 用户鉴权token
}

func NewUserController() *UserController {
	return &UserController{}
}

func (uc *UserController) Info(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &UserInfoRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		response.Error(ctx, response.CodeInvalidParam)
		hlog.Error("controller.UserInfo: 参数校验失败, err: ", err)
		return
	}

	// 从认证中间件中获取userID
	userID := ctx.MustGet(middleware.CtxUserIDKey).(int64)

	// 业务逻辑处理
	resp, err := service.UserInfo(req.UserID, userID)
	if err != nil {
		if errors.Is(err, dao.ErrUserNotExist) {
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

func (uc *UserController) Register(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &UserRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		response.Error(ctx, response.CodeInvalidParam)
		hlog.Error("controller.Register: 参数校验失败, err: ", err)
		return
	}

	// 业务逻辑处理
	resp, err := service.Register(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, dao.ErrUserExist) {
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

func (uc *UserController) Login(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &UserRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		response.Error(ctx, response.CodeInvalidParam)
		hlog.Error("controller.Login: 参数校验失败, err: ", err)
		return
	}

	// 业务逻辑处理
	resp, err := service.Login(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, dao.ErrUserNotExist) {
			response.Error(ctx, response.CodeUserNotExist)
			hlog.Error("controller.Login: 用户不存在")
			return
		}
		if errors.Is(err, dao.ErrPassword) {
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
