package controller

import (
	"context"
	"douyin/config"
	"douyin/dal"
	"douyin/pkg/jwt"
	"douyin/rpc/client"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/kitex/pkg/klog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

type UserController struct{}

type UserInfoRequest struct {
	UserID int64  `query:"user_id,string" vd:"$>0"` // 用户id
	Token  string `query:"token"`                   // 用户登录状态下设置
}

// 用户注册/登陆请求
type UserRequest struct {
	Username string `query:"username" vd:"0<len($)&&len($)<33"` // 注册用户名，最长32个字符
	Password string `query:"password" vd:"5<len($)&&len($)<33"` // 密码，最长32个字符
}

func NewUserController() *UserController {
	return &UserController{}
}

func (uc *UserController) Info(c context.Context, ctx *app.RequestContext) {
	_, span := otel.Tracer(config.Conf.OpenTelemetryConfig.ApiName).Start(c, "controller.UserInfo")
	defer span.End()

	// 获取参数
	req := &UserInfoRequest{}
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
	resp, err := client.UserInfo(req.UserID, userID)
	if err != nil {
		span.RecordError(err)

		if errors.Is(err, dal.ErrUserNotExist) {
			Error(ctx, CodeUserNotExist)
			span.SetStatus(codes.Error, "用户不存在")
			klog.Error("用户不存在")
			return
		}
		Error(ctx, CodeServerBusy)
		span.SetStatus(codes.Error, "业务处理失败")
		klog.Error("业务处理失败, err: ", err)
		return
	}

	// 返回响应
	Success(ctx, resp)
}

func (uc *UserController) Register(c context.Context, ctx *app.RequestContext) {
	_, span := otel.Tracer(config.Conf.OpenTelemetryConfig.ApiName).Start(c, "controller.Register")
	defer span.End()

	// 获取参数
	req := &UserRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
		span.RecordError(err)
		span.SetStatus(codes.Error, "参数校验失败")
		klog.Error("参数校验失败, err: ", err)
		return
	}

	// 业务逻辑处理
	resp, err := client.Register(req.Username, req.Password)
	if err != nil {
		span.RecordError(err)

		if errors.Is(err, dal.ErrUserExist) {
			Error(ctx, CodeUserExist)
			span.SetStatus(codes.Error, "用户已存在")
			klog.Error("用户已存在")
			return
		}
		Error(ctx, CodeServerBusy)
		span.SetStatus(codes.Error, "业务处理失败")
		klog.Error("业务处理失败, err: ", err)
		return
	}

	// 返回响应
	Success(ctx, resp)
}

func (uc *UserController) Login(c context.Context, ctx *app.RequestContext) {
	_, span := otel.Tracer(config.Conf.OpenTelemetryConfig.ApiName).Start(c, "controller.Login")
	defer span.End()

	// 获取参数
	req := &UserRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
		span.RecordError(err)
		span.SetStatus(codes.Error, "参数校验失败")
		klog.Error("参数校验失败, err: ", err)
		return
	}

	// 业务逻辑处理
	resp, err := client.Login(req.Username, req.Password)
	if err != nil {
		span.RecordError(err)

		if errors.Is(err, dal.ErrUserNotExist) {
			Error(ctx, CodeUserNotExist)
			span.SetStatus(codes.Error, "用户不存在")
			klog.Error("用户不存在")
			return
		}
		if errors.Is(err, dal.ErrPassword) {
			Error(ctx, CodeInvalidPassword)
			span.SetStatus(codes.Error, "密码错误")
			klog.Error("密码错误")
			return
		}
		Error(ctx, CodeServerBusy)
		span.SetStatus(codes.Error, "业务处理失败")
		klog.Error("业务处理失败, err: ", err)
		return
	}

	// 返回响应
	Success(ctx, resp)
}
