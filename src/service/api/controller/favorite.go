package controller

import (
	"context"

	"douyin/src/config"
	"douyin/src/dal"
	"douyin/src/kitex_gen/favorite"
	"douyin/src/kitex_gen/favorite/favoriteservice"
	"douyin/src/pkg/jwt"
	"douyin/src/pkg/tracing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	tracing2 "github.com/kitex-contrib/obs-opentelemetry/tracing"
	consul "github.com/kitex-contrib/registry-consul"
	"go.opentelemetry.io/otel/codes"
)

type FavoriteController struct{}

type FavoriteActionRequest struct {
	VideoID    int64 `query:"video_id,string"    vd:"$>0"`        // 视频id
	ActionType int64 `query:"action_type,string" vd:"$==1||$==2"` // 1-点赞，2-取消点赞
}

type FavoriteListRequest struct {
	ToUserID int64  `query:"user_id,string" vd:"$>0"` // 用户id
	Token    string `query:"token"`                   // 用户登录状态下设置
}

var favoriteClient favoriteservice.Client

func init() {
	// 服务发现
	r, err := consul.NewConsulResolver(config.Conf.ConsulConfig.ConsulAddr)
	if err != nil {
		panic(err)
	}

	favoriteClient, err = favoriteservice.NewClient(
		config.Conf.OpenTelemetryConfig.FavoriteName,
		client.WithResolver(r),
		client.WithSuite(tracing2.NewClientSuite()),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: config.Conf.OpenTelemetryConfig.ApiName}),
		client.WithMuxConnection(2),
	)
	if err != nil {
		panic(err)
	}
}

func NewFavoriteController() *FavoriteController {
	return &FavoriteController{}
}

func (fc *FavoriteController) Action(c context.Context, ctx *app.RequestContext) {
	c, span := tracing.Tracer.Start(c, "FavoriteAction")
	defer span.End()

	// 获取参数
	req := &FavoriteActionRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		Error(ctx, CodeInvalidParam)
		span.RecordError(err)
		span.SetStatus(codes.Error, "参数校验失败")
		hlog.Error("参数校验失败, err: ", err)
		return
	}

	// 解析视频点赞类型
	if req.ActionType == 2 {
		req.ActionType = -1
	}

	// 从认证中间件中获取userID
	userID := ctx.MustGet(CtxUserIDKey).(int64)

	// 业务逻辑处理
	resp, err := favoriteClient.FavoriteAction(c, &favorite.FavoriteActionRequest{
		UserId:     userID,
		VideoId:    req.VideoID,
		ActionType: req.ActionType,
	})
	if err != nil {
		span.RecordError(err)
		if errorIs(err, dal.ErrAlreadyFavorite) {
			Error(ctx, CodeAlreadyFavorite)
			span.SetStatus(codes.Error, "已经点赞过了")
			hlog.Error("已经点赞过了")
			return
		}
		if errorIs(err, dal.ErrNotFavorite) {
			Error(ctx, CodeNotFavorite)
			span.SetStatus(codes.Error, "还没有点赞过")
			hlog.Error("还没有点赞过")
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

func (fc *FavoriteController) List(c context.Context, ctx *app.RequestContext) {
	c, span := tracing.Tracer.Start(c, "FavoriteList")
	defer span.End()

	// 获取参数
	req := &FavoriteListRequest{}
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
	resp, err := favoriteClient.FavoriteList(c, &favorite.FavoriteListRequest{
		UserId:   userID,
		ToUserId: req.ToUserID,
	})
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