package controller

import (
	"context"
	"mime/multipart"
	"net/http"

	"douyin/src/pkg/jwt"
	"douyin/src/pkg/tracing"

	"douyin/src/config"
	"douyin/src/kitex_gen/publish"
	"douyin/src/kitex_gen/publish/publishservice"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"go.opentelemetry.io/otel/codes"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	tracing2 "github.com/kitex-contrib/obs-opentelemetry/tracing"
	consul "github.com/kitex-contrib/registry-consul"
)

const (
	minFileSize = 1 * 1024 * 1024   // 1MB
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

var publishClient publishservice.Client

func init() {
	// 服务发现
	r, err := consul.NewConsulResolver(config.Conf.ConsulConfig.ConsulAddr)
	if err != nil {
		panic(err)
	}

	publishClient, err = publishservice.NewClient(
		config.Conf.OpenTelemetryConfig.PublishName,
		client.WithResolver(r),
		client.WithSuite(tracing2.NewClientSuite()),
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: config.Conf.OpenTelemetryConfig.ApiName}),
		client.WithMuxConnection(2),
	)
	if err != nil {
		panic(err)
	}
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

	// 打开文件
	file, err := req.Data.Open()
	if err != nil {
		Error(ctx, CodeServerBusy)
		span.RecordError(err)
		span.SetStatus(codes.Error, "文件打开失败")
		hlog.Error("文件打开失败, err: ", err)
		return
	}
	defer file.Close()

	// 将文件转换为[]byte
	buf := make([]byte, req.Data.Size)
	if _, err := file.Read(buf); err != nil {
		Error(ctx, CodeServerBusy)
		span.RecordError(err)
		span.SetStatus(codes.Error, "文件读取失败")
		hlog.Error("文件读取失败, err: ", err)
		return
	}

	// 判断文件MIME类型是否是视频
	contentType := http.DetectContentType(buf)
	if contentType[:5] != "video" {
		Error(ctx, CodeInvalidParam)
		hlog.Warn("文件类型不是视频")
		return
	}

	// 从认证中间件中获取userID
	userID := ctx.MustGet(CtxUserIDKey).(int64)

	// 业务逻辑处理
	resp, err := publishClient.PublishAction(c, &publish.PublishActionRequest{
		UserId: userID,
		Data:   buf,
		Title:  req.Title,
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
	resp, err := publishClient.PublishList(c, &publish.PublishListRequest{
		UserId:   userID,
		ToUserId: authorID,
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
