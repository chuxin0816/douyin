package controller

import (
	"context"

	"douyin/src/dal"
	"douyin/src/pkg/jwt"
	"douyin/src/pkg/tracing"

	"douyin/src/config"
	"douyin/src/kitex_gen/user"
	"douyin/src/kitex_gen/user/userservice"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"go.opentelemetry.io/otel/codes"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	tracing2 "github.com/kitex-contrib/obs-opentelemetry/tracing"
	consul "github.com/kitex-contrib/registry-consul"
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

var userClient userservice.Client

func init() {
	// 服务发现
	r, err := consul.NewConsulResolver(config.Conf.ConsulConfig.ConsulAddr)
	if err != nil {
		panic(err)
	}

	userClient, err = userservice.NewClient(
		config.Conf.OpenTelemetryConfig.UserName,
		client.WithResolver(r),
		client.WithSuite(tracing2.NewClientSuite()),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: config.Conf.OpenTelemetryConfig.ApiName}),
		client.WithMuxConnection(2),
	)
	if err != nil {
		panic(err)
	}
}

func NewUserController() *UserController {
	return &UserController{}
}

func (uc *UserController) Info(c context.Context, ctx *app.RequestContext) {
	c, span := tracing.Tracer.Start(c, "UserInfo")
	defer span.End()

	// 获取参数
	req := &UserInfoRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
		span.RecordError(err)
		span.SetStatus(codes.Error, "参数校验失败")
		hlog.Error("参数校验失败, err: ", err)
		return
	}

	// 验证token
	userID := jwt.ParseToken(req.Token)

	// 业务逻辑处理
	resp, err := userClient.UserInfo(c, &user.UserInfoRequest{
		UserId:   userID,
		ToUserId: req.UserID,
	})
	if err != nil {
		span.RecordError(err)

		if errorIs(err, dal.ErrUserNotExist) {
			Error(ctx, CodeUserNotExist)
			span.SetStatus(codes.Error, "用户不存在")
			hlog.Error("用户不存在")
			return
		}
		Error(ctx, CodeServerBusy)
		span.SetStatus(codes.Error, "业务处理失败")
		hlog.Error("业务处理失败, err: ", err)
		return
	}

	// 返回响应
	Success(ctx, resp)
}

func (uc *UserController) Register(c context.Context, ctx *app.RequestContext) {
	c, span := tracing.Tracer.Start(c, "Register")
	defer span.End()

	// 获取参数
	req := &UserRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
		span.RecordError(err)
		span.SetStatus(codes.Error, "参数校验失败")
		hlog.Error("参数校验失败, err: ", err)
		return
	}

	// 业务逻辑处理
	resp, err := userClient.Register(c, &user.UserRegisterRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		span.RecordError(err)

		if errorIs(err, dal.ErrUserExist) {
			Error(ctx, CodeUserExist)
			span.SetStatus(codes.Error, "用户已存在")
			hlog.Error("用户已存在")
			return
		}
		Error(ctx, CodeServerBusy)
		span.SetStatus(codes.Error, "业务处理失败")
		hlog.Error("业务处理失败, err: ", err)
		return
	}

	// 返回响应
	Success(ctx, resp)
}

func (uc *UserController) Login(c context.Context, ctx *app.RequestContext) {
	c, span := tracing.Tracer.Start(c, "Login")
	defer span.End()

	// 获取参数
	req := &UserRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
		span.RecordError(err)
		span.SetStatus(codes.Error, "参数校验失败")
		hlog.Error("参数校验失败, err: ", err)
		return
	}

	// 业务逻辑处理
	resp, err := userClient.Login(c, &user.UserLoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		span.RecordError(err)

		if errorIs(err, dal.ErrUserNotExist) {
			Error(ctx, CodeUserNotExist)
			span.SetStatus(codes.Error, "用户不存在")
			hlog.Error("用户不存在")
			return
		}
		if errorIs(err, dal.ErrPassword) {
			Error(ctx, CodeInvalidPassword)
			span.SetStatus(codes.Error, "密码错误")
			hlog.Error("密码错误")
			return
		}
		Error(ctx, CodeServerBusy)
		span.SetStatus(codes.Error, "业务处理失败")
		hlog.Error("业务处理失败, err: ", err)
		return
	}

	// 返回响应
	Success(ctx, resp)
}