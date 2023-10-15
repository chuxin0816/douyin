package controller

import (
	"context"
	"douyin/models"
	"douyin/pkg/jwt"
	"douyin/response"
	"douyin/service"
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type PublishController struct{}

func NewPublishController() *PublishController {
	return &PublishController{}
}

func (pc *PublishController) Action(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	req := &models.ActionRequest{}
	err := ctx.BindAndValidate(req)
	if err != nil {
		response.Error(ctx, response.CodeInvalidParam)
		hlog.Error("controller.Action: 参数校验失败, err: ", err)
		return
	}

	// 验证大小
	if req.Data.Size > 1024*1024*100 {
		response.Error(ctx, response.CodeInvalidParam)
		hlog.Error("controller.Action: 文件太大")
		return
	}

	// 验证token
	userID, err := jwt.ParseToken(req.Token)
	if err != nil {
		if errors.Is(err, jwt.ErrInvalidToken) {
			response.Error(ctx, response.CodeNoAuthority)
			hlog.Error("controller.Action: token无效")
			return
		}
		response.Error(ctx, response.CodeServerBusy)
		hlog.Error("controller.Action: jwt解析出错, err: ", err)
		return
	}

	// 业务逻辑处理
	resp, err := service.PublishAction(ctx, userID, req.Data, req.Title)
	if err != nil {
		response.Error(ctx, response.CodeServerBusy)
		hlog.Error("controller.Action: 业务处理失败, err: ", err)
		return
	}

	// 返回响应
	response.Success(ctx, resp)
}

func (pc *PublishController) List(c context.Context, ctx *app.RequestContext) {
	// 获取参数
	// req := &models.ListRequest{}
}
